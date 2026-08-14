// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"dhcli/handlers/utils"
	"dhcli/keys"

	"github.com/spf13/viper"
)

//go:embed proxy_page.html
var proxyPageFS embed.FS

const browserProxyTimeout = 2 * time.Minute

// proxyPageData holds the template data for the authentication bootstrap page.
type proxyPageData struct {
	AuthEndpoint string
	Token        string
	ServiceURL   string
}

// StartBrowserProxy bootstraps an authenticated browser session to a remote
// service. It starts a temporary localhost HTTP server, serves a single
// auto-submitting HTML form, and opens the browser. The form POSTs credentials
// to the remote proxy's /auth endpoint, which sets a session cookie and
// redirects the browser directly to the service. The local server shuts down
// once the page has been delivered or after a timeout.
//
// Unlike port-forward, the CLI does NOT proxy any HTTP traffic. After the
// initial bootstrap the browser communicates directly with the remote proxy.
func StartBrowserProxy(ctx context.Context, project string, runID string) error {
	logger := utils.GetGlobalLogger()

	authToken := viper.GetString(keys.DhCoreAccessToken)
	if authToken == "" {
		return fmt.Errorf("authorization token not available")
	}

	proxyURLStr := viper.GetString(keys.DhCoreProxy)
	if proxyURLStr == "" {
		return fmt.Errorf("proxy URL not configured")
	}

	// Resolve service base URL from the run resource
	logger.Step("Resolving service URL...")
	service := &ServiceInfo{}
	if err := refreshServiceInfo(service, project, runID); err != nil {
		return err
	}

	authEndpoint := strings.TrimSuffix(proxyURLStr, "/") + "/auth"

	logger.Step("Authenticating browser...")
	localURL, doneCh, stop, err := startBrowserProxyServer(ctx, authEndpoint, authToken, service.BaseURL)
	if err != nil {
		return fmt.Errorf("failed to start local server: %w", err)
	}
	defer stop()

	logger.Step("Opening browser...")
	if err := utils.OpenBrowser(localURL); err != nil {
		logger.Error(fmt.Sprintf("failed to open browser: %v", err))
		logger.Info(fmt.Sprintf("Open manually: %s", localURL))
	}

	logger.Info("Waiting for browser connection...")

	select {
	case <-doneCh:
		// page was served; browser is now navigating to the remote service
	case <-ctx.Done():
		// cancelled by signal or parent context
	}

	return nil
}

// startBrowserProxyServer starts a temporary localhost HTTP server that serves
// the authentication bootstrap page. It returns the local URL to open, a done
// channel that is closed once the page has been served (or the timeout fires),
// a stop function, and any startup error.
func startBrowserProxyServer(
	ctx context.Context,
	authEndpoint string,
	authToken string,
	serviceBaseURL string,
) (localURL string, doneCh <-chan struct{}, stop func(), err error) {
	logger := utils.GetGlobalLogger()

	ctx, cancel := context.WithTimeout(ctx, browserProxyTimeout)

	tmpl, tmplErr := template.ParseFS(proxyPageFS, "proxy_page.html")
	if tmplErr != nil {
		cancel()
		return "", nil, nil, fmt.Errorf("failed to parse proxy page template: %w", tmplErr)
	}

	mux := http.NewServeMux()
	srv := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      15 * time.Second,
		ReadTimeout:       15 * time.Second,
	}

	ln, listenErr := net.Listen("tcp", ":0")
	if listenErr != nil {
		cancel()
		return "", nil, nil, fmt.Errorf("failed to bind local port: %w", listenErr)
	}

	port := ln.Addr().(*net.TCPAddr).Port
	url := fmt.Sprintf("http://localhost:%d/", port)

	done := make(chan struct{})
	doneOnce := sync.Once{}
	closeDone := func() { doneOnce.Do(func() { close(done) }) }

	stopOnce := sync.Once{}
	stopFn := func() {
		stopOnce.Do(func() {
			cancel()
			closeDone()
			shutdownCtx, c := context.WithTimeout(context.Background(), 3*time.Second)
			defer c()
			_ = srv.Shutdown(shutdownCtx)
		})
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Ignore requests that arrive after shutdown has been initiated
		select {
		case <-ctx.Done():
			http.Error(w, "gone", http.StatusGone)
			return
		default:
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")

		data := proxyPageData{
			AuthEndpoint: authEndpoint,
			Token:        authToken,
			ServiceURL:   serviceBaseURL,
		}

		if execErr := tmpl.Execute(w, data); execErr != nil {
			logger.Error(fmt.Sprintf("proxy page template error: %v", execErr))
		}

		logger.Info("Authentication page served.")

		// Shut down after delivering the page. Use a goroutine so the HTTP
		// response is fully flushed before the server is closed.
		go stopFn()
	})

	ready := make(chan struct{})
	go func() {
		close(ready)
		_ = srv.Serve(ln)
	}()
	<-ready

	// Auto-stop when the timeout fires
	go func() {
		<-ctx.Done()
		stopFn()
	}()

	return url, done, stopFn, nil
}
