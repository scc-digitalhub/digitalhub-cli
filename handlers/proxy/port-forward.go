package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
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
	dt.logger.Debug(fmt.Sprintf("✈️ [Port-Forward] %s %s", req.Method, req.URL.String()))
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

// ServiceInfo holds the resolved run information
type ServiceInfo struct {
	BaseURL   string
	FetchedAt time.Time
	Host      string
}

// portForwardTargetKey is the context key used to pass pre-computed target
// information from the outer handler into the ReverseProxy Rewrite hook.
type portForwardTargetKey struct{}

// portForwardTarget holds per-request target information derived from the
// current ServiceInfo, pre-computed before the ReverseProxy takes over.
type portForwardTarget struct {
	basePath string // TrimSuffix(baseURL.Path, "/")
	host     string // hostname[:port] of the service (for X-Proxy-Host)
}

// StartPortForward starts a local HTTP port-forward on the given port that
// tunnels traffic to the remote service resolved from the run resource.
// Authorization and routing headers are injected automatically.
// HTTP upgrades (WebSocket) are transparently tunneled.
// If localPort is 0, a random port is assigned.
func StartPortForward(ctx context.Context, project string, runID string, localPort int) error {
	logger := utils.GetGlobalLogger()

	// Get remote proxy URL from configuration
	proxyURLStr := viper.GetString(utils.DhCoreProxy)
	if proxyURLStr == "" {
		return fmt.Errorf("proxy URL not configured")
	}

	proxyURL, err := url.Parse(proxyURLStr)
	if err != nil {
		return fmt.Errorf("invalid proxy URL: %w", err)
	}

	logger.Step(fmt.Sprintf("Using proxy %s", proxyURL.String()))

	// Get authorization token
	authToken := viper.GetString(utils.DhCoreAccessToken)
	if authToken == "" {
		return fmt.Errorf("authorization token not available")
	}

	// Resolved host header value for the remote proxy (constant for lifetime)
	proxyHost := proxyURL.Hostname()
	if proxyURL.Port() != "" {
		proxyHost = proxyURL.Hostname() + ":" + proxyURL.Port()
	}

	// ServiceInfo cache, guarded by mu
	service := &ServiceInfo{}
	var mu sync.Mutex

	if err := refreshServiceInfo(service, project, runID); err != nil {
		return err
	}

	transport := &http.Transport{
		// IMPORTANT for streaming / SSE
		DisableCompression:    true,
		ForceAttemptHTTP2:     false,
		ResponseHeaderTimeout: 0,
	}

	var httpTransport http.RoundTripper = transport

	// Wrap with debug transport if in verbose mode
	if logger.IsVerbose() {
		httpTransport = &debugTransport{
			transport: transport,
			logger:    logger,
		}
	}

	// ReverseProxy performs the actual forwarding, including HTTP Upgrade
	// (WebSocket) tunneling when the upstream responds with 101.
	rp := &httputil.ReverseProxy{
		// Rewrite is called by ReverseProxy for every request (including
		// WebSocket upgrades). It sets the outbound URL and injects the
		// required headers. Target information was pre-computed by the outer
		// handler and stored in the request context.
		Rewrite: func(pr *httputil.ProxyRequest) {
			target := pr.In.Context().Value(portForwardTargetKey{}).(portForwardTarget)

			pr.Out.URL.Scheme = proxyURL.Scheme
			pr.Out.URL.Host = proxyHost
			pr.Out.URL.Path = target.basePath + pr.In.URL.Path
			pr.Out.URL.RawQuery = pr.In.URL.RawQuery

			// Rewrite the Host header so the remote proxy routes correctly
			pr.Out.Host = proxyHost

			// Tell the remote proxy which backend service to reach
			pr.Out.Header.Set("X-Proxy-Host", target.host)

			// Authenticate with the remote proxy
			pr.Out.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
		},
		Transport: httpTransport,
		// FlushInterval -1 means flush immediately after each write, which is
		// required for streaming responses (SSE, chunked transfer, etc.).
		FlushInterval: -1,
	}

	// Outer handler: refresh ServiceInfo cache if stale, parse the base URL
	// (returning a proper HTTP error on failure), then hand off to ReverseProxy.
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		if time.Since(service.FetchedAt) > cacheRefreshInterval {
			if err := refreshServiceInfo(service, project, runID); err != nil {
				mu.Unlock()
				http.Error(w, fmt.Sprintf("Failed to refresh run info: %v", err), 502)
				return
			}
		}
		svcInfo := *service // copy under lock
		mu.Unlock()

		baseURLParsed, err := url.Parse(svcInfo.BaseURL)
		if err != nil {
			http.Error(w, fmt.Sprintf("invalid service base URL: %v", err), 500)
			return
		}

		target := portForwardTarget{
			basePath: strings.TrimSuffix(baseURLParsed.Path, "/"),
			host:     svcInfo.Host,
		}
		r = r.WithContext(context.WithValue(r.Context(), portForwardTargetKey{}, target))
		rp.ServeHTTP(w, r)
	})

	server := &http.Server{
		Handler:      handler,
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

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	port := ln.Addr().(*net.TCPAddr).Port
	logger.Success(fmt.Sprintf("Port-forward listening on localhost:%d", port))
	logger.Info(fmt.Sprintf("Run ID: %s -> Base URL: %s", runID, service.BaseURL))
	logger.Info(fmt.Sprintf("Configure clients to use http://localhost:%d", port))

	// Handle context cancellation
	go func() {
		<-ctx.Done()
		logger.Info("Shutting down port-forward...")
		server.Shutdown(context.Background())
	}()

	return server.Serve(ln)
}

// refreshServiceInfo fetches the run resource and extracts the baseURL
func refreshServiceInfo(service *ServiceInfo, project string, runID string) error {
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

	service.BaseURL = baseURL
	service.Host = hostHeader
	service.FetchedAt = time.Now()

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
