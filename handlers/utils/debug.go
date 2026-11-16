// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

// DebugTransport wraps http.RoundTripper to log requests and responses
type DebugTransport struct {
	underlying http.RoundTripper
}

// RoundTrip implements http.RoundTripper interface
func (t *DebugTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	logger := GetGlobalLogger()

	// Headers, query strings and bodies can all contain credentials.
	logger.Debug(fmt.Sprintf("===== REQUEST =====\n%s %s://%s%s", req.Method, req.URL.Scheme, req.URL.Host, req.URL.EscapedPath()))

	resp, err := t.underlying.RoundTrip(req)
	if err != nil {
		// url.Error embeds the full URL (incl. query), so only log the underlying cause.
		cause := err
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			cause = urlErr.Err
		}
		logger.Debug(fmt.Sprintf("HTTP request failed: %v", cause))
		return resp, err
	}
	logger.Debug(fmt.Sprintf("===== RESPONSE =====\n%s %d", resp.Proto, resp.StatusCode))
	return resp, nil
}

// CreateDebugHTTPClient creates an HTTP client with request/response logging
func CreateDebugHTTPClient() *http.Client {
	return &http.Client{
		Transport: &DebugTransport{
			underlying: http.DefaultTransport,
		},
	}
}

// Global debug HTTP client
var debugHTTPClient *http.Client

// EnableDebugMode enables HTTP debug logging and returns the debug client
func EnableDebugMode() *http.Client {
	if debugHTTPClient == nil {
		debugHTTPClient = CreateDebugHTTPClient()
	}
	return debugHTTPClient
}

// GetDebugHTTPClient returns the global debug HTTP client if debug mode is enabled
func GetDebugHTTPClient() *http.Client {
	return debugHTTPClient
}

// SetHTTPClient sets the global HTTP client used by the CLI and SDK (useful for testing or custom transports).
func SetHTTPClient(client *http.Client) {
	debugHTTPClient = client
}
