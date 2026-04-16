package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strings"
	"time"

	"dhcli/handlers/utils"

	"github.com/scc-digitalhub/digitalhub-cli-sdk/sdk/config"
	crudsvc "github.com/scc-digitalhub/digitalhub-cli-sdk/sdk/services/crud"
	"github.com/spf13/viper"
)

const (
	cacheRefreshInterval = 2 * time.Minute
)

// debugTransport wraps an http.Transport with debug logging
type debugTransport struct {
	transport *http.Transport
	logger    *utils.StepLogger
}

func (dt *debugTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	dt.logger.Debug(fmt.Sprintf("✈️ [Proxy] %s %s", req.Method, req.URL.String()))
	dt.logger.Debug(fmt.Sprintf("   Headers: %v", req.Header))

	// Add trace to log connection events
	trace := &httptrace.ClientTrace{
		ConnectStart: func(network, addr string) {
			dt.logger.Debug(fmt.Sprintf("   Connecting to %s (%s)", addr, network))
		},
		ConnectDone: func(network, addr string, err error) {
			if err != nil {
				dt.logger.Debug(fmt.Sprintf("   Connection failed: %v", err))
			} else {
				dt.logger.Debug(fmt.Sprintf("   Connected to %s", addr))
			}
		},
		GotFirstResponseByte: func() {
			dt.logger.Debug("   Received first response byte")
		},
	}
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

	resp, err := dt.transport.RoundTrip(req)
	if err != nil {
		dt.logger.Debug(fmt.Sprintf("   ❌ Error: %v", err))
		return resp, err
	}

	dt.logger.Debug(fmt.Sprintf("   Response: %d %s", resp.StatusCode, resp.Status))
	dt.logger.Debug(fmt.Sprintf("   Response Headers: %v", resp.Header))

	return resp, nil
}

func (dt *debugTransport) CloseIdleConnections() {
	dt.transport.CloseIdleConnections()
}

// RunInfo holds the resolved run information
type RunInfo struct {
	BaseURL    string
	FetchedAt  time.Time
	HostHeader string
}

// StartProxy starts a transparent HTTP proxy on a local port
// that forwards requests to a baseUrl resolved from a run resource
// If localPort is 0, a random port will be assigned
func StartProxy(ctx context.Context, project string, runID string, localPort int) error {
	// Get proxy configuration from viper
	proxyURLStr := viper.GetString(utils.DhCoreProxy)
	if proxyURLStr == "" {
		return fmt.Errorf("proxy URL not configured")
	}

	proxyURL, err := url.Parse(proxyURLStr)
	if err != nil {
		return fmt.Errorf("invalid proxy URL: %w", err)
	}

	// Get authorization token
	authToken := viper.GetString(utils.DhCoreAccessToken)
	if authToken == "" {
		return fmt.Errorf("authorization token not available")
	}

	// Initialize run info cache
	runInfo := &RunInfo{}

	// Fetch run info immediately
	if err := refreshRunInfo(runInfo, project, runID); err != nil {
		return err
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),

		// IMPORTANT for streaming
		DisableCompression:    true,
		ForceAttemptHTTP2:     false,
		ResponseHeaderTimeout: 0,
	}

	var httpTransport http.RoundTripper = transport

	// Wrap with debug transport if in verbose mode
	logger := utils.GetGlobalLogger()
	if logger.IsVerbose() {
		httpTransport = &debugTransport{
			transport: transport,
			logger:    logger,
		}
	}

	client := &http.Client{
		Transport: httpTransport,
		Timeout:   0, // no timeout for streaming
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		// Refresh run info if cache is stale
		if time.Since(runInfo.FetchedAt) > cacheRefreshInterval {
			if err := refreshRunInfo(runInfo, project, runID); err != nil {
				http.Error(w, fmt.Sprintf("Failed to refresh run info: %v", err), 502)
				return
			}
		}

		// Build target URL using the resolved baseURL
		targetURL := fmt.Sprintf("%s%s", runInfo.BaseURL, r.URL.Path)
		if r.URL.RawQuery != "" {
			targetURL += "?" + r.URL.RawQuery
		}

		req, err := http.NewRequest(r.Method, targetURL, r.Body)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		req.Header = r.Header.Clone()

		// Use the resolved hostname
		req.Header.Set("Host", runInfo.HostHeader)

		// Inject Authorization for remote proxy
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, err.Error(), 502)
			return
		}
		defer resp.Body.Close()

		// Copy headers
		for k, vv := range resp.Header {
			for _, v := range vv {
				w.Header().Add(k, v)
			}
		}

		w.WriteHeader(resp.StatusCode)

		// 🔥 STREAMING: copy as-is
		flusher, ok := w.(http.Flusher)

		buf := make([]byte, 32*1024)
		for {
			n, err := resp.Body.Read(buf)
			if n > 0 {
				w.Write(buf[:n])
				if ok {
					flusher.Flush() // push chunks immediately
				}
			}
			if err != nil {
				if err != io.EOF {
					logger := utils.GetGlobalLogger()
					logger.Debug(fmt.Sprintf("stream error: %v", err))
				}
				break
			}
		}
	}

	server := &http.Server{
		Addr:         ":0", // random port
		Handler:      http.HandlerFunc(handler),
		ReadTimeout:  0,
		WriteTimeout: 0,
		IdleTimeout:  0,
		BaseContext: func(l net.Listener) context.Context {
			return ctx
		},
	}

	// Build address string based on localPort
	var addr string
	if localPort == 0 {
		addr = ":0" // random port
	} else {
		addr = fmt.Sprintf(":%d", localPort)
	}

	// Listen on specified or random port
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	port := ln.Addr().(*net.TCPAddr).Port
	logger.Success(fmt.Sprintf("Transparent proxy listening on localhost:%d", port))
	logger.Info(fmt.Sprintf("Run ID: %s -> Base URL: %s", runID, runInfo.BaseURL))
	logger.Info(fmt.Sprintf("Configure clients to use http://localhost:%d", port))

	// Handle context cancellation
	go func() {
		<-ctx.Done()
		logger.Info("Shutting down proxy...")
		server.Shutdown(context.Background())
	}()

	return server.Serve(ln)
}

// refreshRunInfo fetches the run resource and extracts the baseURL
func refreshRunInfo(runInfo *RunInfo, project string, runID string) error {
	logger := utils.GetGlobalLogger()
	logger.Debug(fmt.Sprintf("Fetching run %s in project %s", runID, project))

	// Build SDK config from viper
	cfg := config.Config{
		Core: config.CoreConfig{
			BaseURL:     viper.GetString(utils.DhCoreEndpoint),
			APIVersion:  viper.GetString(utils.DhCoreApiVersion),
			AccessToken: viper.GetString(utils.DhCoreAccessToken),
		},
		HTTPClient: utils.GetDebugHTTPClient(),
	}

	ctx := context.Background()

	// Create CRUD service
	crud, err := crudsvc.NewCrudService(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to create CRUD service: %w", err)
	}

	// Translate resource name to API endpoint (e.g., "run" -> "runs")
	endpoint := utils.TranslateEndpoint("run")

	// Get the run resource with project
	body, _, err := crud.Get(ctx, crudsvc.GetRequest{
		ResourceRequest: crudsvc.ResourceRequest{
			Project:  project,
			Resource: endpoint,
		},
		ID: runID,
	})
	if err != nil {
		return fmt.Errorf("failed to fetch run %s in project %s: %w", runID, project, err)
	}

	// Extract baseURL from .status.service.baseUrl
	var runData map[string]interface{}
	if err := json.Unmarshal(body, &runData); err != nil {
		return fmt.Errorf("failed to parse run response: %w", err)
	}

	baseURL, err := extractBaseURL(runData)
	if err != nil {
		return err
	}

	// Parse URL to extract hostname
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return fmt.Errorf("invalid baseUrl: %w", err)
	}

	hostHeader := parsedURL.Hostname()
	if parsedURL.Port() != "" {
		hostHeader = parsedURL.Hostname() + ":" + parsedURL.Port()
	}

	runInfo.BaseURL = baseURL
	runInfo.HostHeader = hostHeader
	runInfo.FetchedAt = time.Now()

	logger.Debug(fmt.Sprintf("Run baseURL resolved to: %s (host: %s)", baseURL, hostHeader))

	return nil
}

// extractBaseURL extracts the baseUrl from the run resource structure
func extractBaseURL(data map[string]interface{}) (string, error) {
	status, ok := data["status"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("missing or invalid .status in run resource")
	}

	service, ok := status["service"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("missing or invalid .status.service in run resource")
	}

	baseURL, ok := service["url"].(string)
	if !ok || baseURL == "" {
		return "", fmt.Errorf("missing or empty .status.service.url in run resource")
	}

	// If url doesn't include protocol, add http://
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "http://" + baseURL
	}

	return baseURL, nil
}
