// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-stomp/stomp/v3"
	"github.com/gorilla/websocket"

	"dhcli/handlers/utils"
	"dhcli/keys"

	"github.com/spf13/viper"
)

// wsNetConn wraps a gorilla WebSocket connection so it satisfies net.Conn,
// which is what go-stomp/stomp expects.
type wsNetConn struct {
	conn    *websocket.Conn
	readBuf []byte
	debug   bool
}

func (w *wsNetConn) Read(p []byte) (int, error) {
	for len(w.readBuf) == 0 {
		_, msg, err := w.conn.ReadMessage()
		if err != nil {
			return 0, err
		}
		if w.debug {
			utils.GetGlobalLogger().Debug("[STOMP] <<<\n" + string(msg))
		}
		w.readBuf = msg
	}
	n := copy(p, w.readBuf)
	w.readBuf = w.readBuf[n:]
	return n, nil
}

func (w *wsNetConn) Write(p []byte) (int, error) {
	if w.debug {
		utils.GetGlobalLogger().Debug("[STOMP] >>>\n" + string(p))
	}
	err := w.conn.WriteMessage(websocket.BinaryMessage, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func (w *wsNetConn) Close() error                       { return w.conn.Close() }
func (w *wsNetConn) LocalAddr() net.Addr                { return w.conn.LocalAddr() }
func (w *wsNetConn) RemoteAddr() net.Addr               { return w.conn.RemoteAddr() }
func (w *wsNetConn) SetDeadline(t time.Time) error      { return w.conn.SetReadDeadline(t) }
func (w *wsNetConn) SetReadDeadline(t time.Time) error  { return w.conn.SetReadDeadline(t) }
func (w *wsNetConn) SetWriteDeadline(t time.Time) error { return w.conn.SetWriteDeadline(t) }

// stompLogger bridges the stomp.Logger interface to dhcli's global logger so
// that STOMP library-internal errors appear in debug output when --debug is set.
type stompLogger struct{ l *utils.StepLogger }

func (s stompLogger) Debugf(f string, v ...interface{}) { s.l.Debug(fmt.Sprintf("[STOMP] "+f, v...)) }
func (s stompLogger) Infof(f string, v ...interface{})  { s.l.Debug(fmt.Sprintf("[STOMP] "+f, v...)) }
func (s stompLogger) Warningf(f string, v ...interface{}) {
	s.l.Debug(fmt.Sprintf("[STOMP] WARN "+f, v...))
}
func (s stompLogger) Errorf(f string, v ...interface{}) {
	s.l.Debug(fmt.Sprintf("[STOMP] ERR "+f, v...))
}
func (s stompLogger) Debug(m string)   { s.l.Debug("[STOMP] " + m) }
func (s stompLogger) Info(m string)    { s.l.Debug("[STOMP] " + m) }
func (s stompLogger) Warning(m string) { s.l.Debug("[STOMP] WARN " + m) }
func (s stompLogger) Error(m string)   { s.l.Debug("[STOMP] ERR " + m) }

// EventsHandler connects to the STOMP broker via WebSocket and streams
// push-notifications for the given resource (and optionally a specific ID).
func EventsHandler(env string, output string, project string, name string, resource string, id string) error {
	utils.CheckUpdateEnvironment()
	utils.CheckApiLevel(keys.ApiLevelKey, keys.EventsMin, keys.EventsMax)
	if err := utils.CheckCredentials(); err != nil {
		return err
	}

	endpoint := utils.TranslateEndpoint(resource)
	format := utils.TranslateFormat(output)

	baseURL := viper.GetString(keys.DhCoreEndpoint)
	accessToken := viper.GetString(keys.DhCoreAccessToken)

	wsURL, err := buildWSURL(baseURL)
	if err != nil {
		return fmt.Errorf("could not derive WebSocket URL from endpoint %q: %w", baseURL, err)
	}

	// Build the STOMP subscription destination.
	destination := "/notifications/" + endpoint
	if id != "" {
		destination += "/" + id
	}

	// Dial WebSocket.
	dialer := websocket.DefaultDialer
	wsConn, _, err := dialer.Dial(wsURL, http.Header{})
	if err != nil {
		return fmt.Errorf("WebSocket dial failed: %w", err)
	}
	netConn := &wsNetConn{conn: wsConn, debug: utils.GetDebugHTTPClient() != nil}

	// STOMP connect.
	stompOpts := []func(*stomp.Conn) error{
		stomp.ConnOpt.Header("Authorization", "Bearer "+accessToken),
		stomp.ConnOpt.AcceptVersion(stomp.V12),
		stomp.ConnOpt.HeartBeat(10*time.Second, 10*time.Second),
	}
	if utils.GetDebugHTTPClient() != nil {
		stompOpts = append(stompOpts, stomp.ConnOpt.Logger(stompLogger{utils.GetGlobalLogger()}))
	}
	stompConn, err := stomp.Connect(netConn, stompOpts...)
	if err != nil {
		wsConn.Close()
		return fmt.Errorf("STOMP connect failed: %w", err)
	}
	defer stompConn.Disconnect()

	// Subscribe.
	sub, err := stompConn.Subscribe(destination, stomp.AckAuto)
	if err != nil {
		return fmt.Errorf("STOMP subscribe to %q failed: %w", destination, err)
	}
	defer sub.Unsubscribe()

	fmt.Fprintf(os.Stderr, "Subscribed to %s — waiting for events (Ctrl+C to stop)\n", destination)

	// Closing the raw WebSocket unblocks the STOMP reader goroutine, which
	// closes sub.C and lets the range loop below exit cleanly.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		signal.Stop(sigCh)
		wsConn.Close()
	}()

	for msg := range sub.C {
		if msg.Err != nil {
			// Normal on shutdown; suppress the noise.
			return nil
		}

		var event map[string]interface{}
		if err := json.Unmarshal(msg.Body, &event); err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not parse message: %v\n", err)
			continue
		}

		// Client-side filters.
		if project != "" && !eventMatchesProject(event, project) {
			continue
		}
		if name != "" && !eventMatchesName(event, name) {
			continue
		}

		if err := printEvent(event, format); err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not render event: %v\n", err)
		}
	}
	return nil
}

// buildWSURL converts an http(s) core endpoint into a ws(s)://host/ws URL.
func buildWSURL(baseURL string) (string, error) {
	if baseURL == "" {
		return "", errors.New("endpoint is empty — check your environment config")
	}
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	switch strings.ToLower(u.Scheme) {
	case "https":
		u.Scheme = "wss"
	default:
		u.Scheme = "ws"
	}
	u.Path = "/ws"
	u.RawQuery = ""
	return u.String(), nil
}

// eventMatchesName returns true when the event's record name equals name.
func eventMatchesName(event map[string]interface{}, name string) bool {
	record, ok := event["record"].(map[string]interface{})
	if !ok {
		return false
	}
	n, ok := record["name"].(string)
	return ok && n == name
}

// eventMatchesProject returns true when the event's record belongs to project.
func eventMatchesProject(event map[string]interface{}, project string) bool {
	record, ok := event["record"].(map[string]interface{})
	if !ok {
		return false
	}
	// Check top-level project field on record.
	if p, ok := record["project"].(string); ok && p == project {
		return true
	}
	// Also check metadata.project.
	if meta, ok := record["metadata"].(map[string]interface{}); ok {
		if p, ok := meta["project"].(string); ok && p == project {
			return true
		}
	}
	return false
}

// printEvent marshals the record from the event envelope and delegates to
// the same printShort / printJson / printYaml used by GetHandler.
func printEvent(event map[string]interface{}, format string) error {
	record := eventRecord(event)

	src, err := json.Marshal(record)
	if err != nil {
		return err
	}

	// Pass a non-empty id so printJson/printYaml skip the GetFirstIfList path
	// (the record is already a single object, not a paginated list).
	id := fmt.Sprintf("%v", record["id"])

	switch format {
	case "json":
		return printJson(id, src, false)
	case "yaml":
		fmt.Println("---")
		return printYaml(id, src)
	default:
		err = printShort(src)
		if err == nil {
			fmt.Println("---")
		}
		return err
	}
}

// eventRecord returns the record from the event envelope, or the full event if absent.
func eventRecord(event map[string]interface{}) map[string]interface{} {
	if record, ok := event["record"].(map[string]interface{}); ok {
		return record
	}
	return event
}
