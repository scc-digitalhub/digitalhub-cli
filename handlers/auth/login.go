// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/spf13/viper"

	"dhcli/handlers/utils"
	"dhcli/keys"
)

//go:embed callback.html
var callbackFS embed.FS

type OAuthResult struct {
	TokenJSON   []byte
	RedirectURI string
	Err         error
}

var logger = utils.GetGlobalLogger()

// ==========================
// PUBLIC ENTRY POINT
// ==========================
func LoginHandler() error {
	utils.CheckUpdateEnvironment()
	utils.CheckApiLevel(keys.ApiLevelKey, keys.LoginMin, keys.LoginMax)

	verifier, challenge := generatePKCE()
	state := randomString(32)

	ctx := context.Background()
	timeout := 180 * time.Second

	redirectURI, resultCh, stop, err := startAuthCodeServer(ctx, verifier, state, timeout)
	if err != nil {
		return fmt.Errorf("impossibile avviare il server locale: %w", err)
	}
	defer stop()

	authURL, err := buildAuthURL(challenge, state, redirectURI)
	if err != nil {
		return err
	}

	fmt.Println("────────────────────────────────────────────────────────────")
	fmt.Println("  The following URL will be opened in your browser:        ")
	fmt.Println("────────────────────────────────────────────────────────────")
	fmt.Println(authURL)
	fmt.Println("────────────────────────────────────────────────────────────")
	fmt.Print("Press Enter to continue... ")

	_, _ = bufio.NewReader(os.Stdin).ReadBytes('\n')

	if err := utils.OpenBrowser(authURL); err != nil {
		logger.Error(fmt.Sprintf("browser open error: %v", err))
	}

	// Wait for result (no select loop hacks)
	res := <-resultCh
	if res.Err != nil {
		return res.Err
	}

	// Map token response into Viper and persist to config file
	var m map[string]interface{}
	if err := json.Unmarshal(res.TokenJSON, &m); err != nil {
		logger.Error(fmt.Sprintf("json parse error: %v", err))
	} else {
		for k, v := range m {
			key := k
			if mapped, ok := keys.DhCoreMap[k]; ok {
				key = mapped
			}
			viper.Set(key, fmt.Sprint(v))
		}

		// Persist config keys to ini file
		if err := utils.UpdateIniSectionFromViper(viper.AllKeys()); err != nil {
			logger.Error(fmt.Sprintf("persist error: %v", err))
		}
	}

	logger.Success("Login successful")
	return nil
}

// ==========================
// PAT (NON-INTERACTIVE) FLOW
// ==========================

// PatLoginHandler performs an OAuth2 token exchange (RFC 8693) using a Personal
// Access Token as the subject token. This is suitable for non-interactive
// environments such as CI jobs or containers.
func PatLoginHandler(pat string) error {
	utils.CheckUpdateEnvironment()
	utils.CheckApiLevel(keys.ApiLevelKey, keys.LoginMin, keys.LoginMax)

	tokenURL := viper.GetString(keys.Oauth2TokenEndpoint)
	if tokenURL == "" {
		return fmt.Errorf("oauth2_token_endpoint not configured")
	}
	clientID := viper.GetString(keys.DhCoreClientId)
	if clientID == "" {
		return fmt.Errorf("dhcore_client_id not configured")
	}

	v := url.Values{
		"grant_type":           {"urn:ietf:params:oauth:grant-type:token-exchange"},
		"client_id":            {clientID},
		"subject_token":        {pat},
		"subject_token_type":   {"urn:ietf:params:oauth:token-type:pat"},
		"requested_token_type": {"urn:ietf:params:oauth:token-type:access_token"},
	}

	client := utils.GetDebugHTTPClient()
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}

	resp, err := client.PostForm(tokenURL, v)
	if err != nil {
		return fmt.Errorf("token exchange request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token exchange error: %s - %s", resp.Status, string(body))
	}

	var m map[string]interface{}
	if err := json.Unmarshal(body, &m); err != nil {
		return fmt.Errorf("failed to parse token response: %w", err)
	}

	for k, val := range m {
		key := k
		if mapped, ok := keys.DhCoreMap[k]; ok {
			key = mapped
		}
		viper.Set(key, fmt.Sprint(val))
	}

	if err := utils.UpdateIniSectionFromViper(viper.AllKeys()); err != nil {
		logger.Error(fmt.Sprintf("persist error: %v", err))
	}

	logger.Success("Login successful (PAT token exchange)")
	return nil
}

// ==========================
// PKCE
// ==========================
func generatePKCE() (verifier, challenge string) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~"

	verifier = randomStringCharset(64, charset)

	h := sha256.Sum256([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(h[:])

	return
}

func randomString(n int) string {
	return randomStringCharset(n, "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")
}

func randomStringCharset(n int, charset string) string {
	b := make([]byte, n)
	for i := range b {
		_, _ = rand.Read(b[i : i+1])
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b)
}

// ==========================
// OAUTH SERVER CORE
// ==========================
func startAuthCodeServer(
	ctx context.Context,
	verifier string,
	state string,
	timeout time.Duration,
) (string, <-chan OAuthResult, func(), error) {

	const maxPortRetries = 5

	out := make(chan OAuthResult, 1)

	ctx, cancel := context.WithTimeout(ctx, timeout)

	mux := http.NewServeMux()

	srv := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      15 * time.Second,
		ReadTimeout:       15 * time.Second,
	}

	// -------------------------
	// PORT BINDING (retry safe)
	// -------------------------
	var ln net.Listener
	var err error

	for i := 0; i < maxPortRetries; i++ {
		ln, err = net.Listen("tcp", ":0")
		if err == nil {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	if ln == nil {
		cancel()
		return "", nil, nil, fmt.Errorf("failed to bind local port")
	}

	port := ln.Addr().(*net.TCPAddr).Port
	redirectURI := fmt.Sprintf("http://localhost:%d/callback", port)

	// -------------------------
	// CLEAN STOP FUNCTION
	// -------------------------
	stopOnce := sync.Once{}
	stop := func() {
		stopOnce.Do(func() {
			cancel()

			shutdownCtx, cancel2 := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel2()

			_ = srv.Shutdown(shutdownCtx)
			close(out)
		})
	}

	// -------------------------
	// CALLBACK HANDLER
	// -------------------------
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {

		// avoid late execution after timeout
		select {
		case <-ctx.Done():
			return
		default:
		}

		if r.URL.Query().Get("state") != state {
			http.Error(w, "invalid state", http.StatusBadRequest)
			select {
			case out <- OAuthResult{Err: fmt.Errorf("state mismatch")}:
			default:
			}
			stop()
			return
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "missing code", http.StatusBadRequest)
			return
		}

		token := exchangeAuthCode(
			viper.GetString(keys.Oauth2TokenEndpoint),
			viper.GetString(keys.DhCoreClientId),
			verifier,
			redirectURI,
			code,
		)

		if token == nil {
			http.Error(w, "token exchange failed", http.StatusInternalServerError)
			select {
			case out <- OAuthResult{Err: fmt.Errorf("token exchange failed")}:
			default:
			}
			stop()
			return
		}

		w.Header().Set("Content-Type", "text/html")

		tmpl, err := template.ParseFS(callbackFS, "callback.html")
		if err != nil {
			http.Error(w, fmt.Sprintf("Template error: %v", err), http.StatusInternalServerError)
			return
		}

		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, token, "", "  "); err != nil {
			prettyJSON.Write(token)
		}

		data := map[string]string{
			"TokenData": prettyJSON.String(),
		}

		if err := tmpl.Execute(w, data); err != nil {
			log.Printf("template execute error: %v", err)
		}
		select {
		case out <- OAuthResult{
			TokenJSON:   token,
			RedirectURI: redirectURI,
		}:
		default:
		}

		stop()
	})

	// -------------------------
	// SERVER START (race-safe)
	// -------------------------
	ready := make(chan struct{})

	go func() {
		close(ready)
		_ = srv.Serve(ln)
	}()

	<-ready

	// auto-stop on timeout
	go func() {
		<-ctx.Done()
		stop()
	}()

	return redirectURI, out, stop, nil
}

// ==========================
// TOKEN EXCHANGE
// ==========================
func exchangeAuthCode(tokenURL, clientID, verifier, redirectURI, code string) []byte {
	v := url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {clientID},
		"code_verifier": {verifier},
		"code":          {code},
		"redirect_uri":  {redirectURI},
	}

	// Use debug HTTP client if available, otherwise use default
	client := utils.GetDebugHTTPClient()
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}

	resp, err := client.PostForm(tokenURL, v)
	if err != nil {
		log.Printf("token request error: %v", err)
		return nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		log.Printf("token error: %s - %s", resp.Status, string(body))
		return nil
	}

	return body
}

// ==========================
// AUTH URL BUILDER
// ==========================
func buildAuthURL(chal, state, redirectURI string) (string, error) {
	raw := viper.GetString("scopes_supported")

	var scopes []string
	if raw != "" {
		split := strings.FieldsFunc(raw, func(r rune) bool {
			return r == ',' || r == ' ' || r == '\n' || r == '\t'
		})

		for _, s := range split {
			if s != "" {
				scopes = append(scopes, s)
			}
		}
	}

	scope := strings.Join(scopes, " ")

	v := url.Values{
		"response_type":         {"code"},
		"client_id":             {viper.GetString(keys.DhCoreClientId)},
		"redirect_uri":          {redirectURI},
		"code_challenge":        {chal},
		"code_challenge_method": {"S256"},
		"state":                 {state},
	}

	base := viper.GetString("authorization_endpoint")
	if base == "" {
		return "", fmt.Errorf("authorization_endpoint missing")
	}

	return base + "?" + v.Encode() + "&scope=" + url.QueryEscape(scope), nil
}
