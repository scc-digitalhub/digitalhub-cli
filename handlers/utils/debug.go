// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
)

// DebugTransport wraps http.RoundTripper to log requests and responses
type DebugTransport struct {
	underlying http.RoundTripper
}

// RoundTrip implements http.RoundTripper interface
func (t *DebugTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	logger := GetGlobalLogger()

	// Log request
	reqDump, err := httputil.DumpRequest(req, true)
	if err != nil {
		logger.Debug(fmt.Sprintf("Error dumping request: %v", err))
	} else {
		logger.Debug(fmt.Sprintf("===== REQUEST =====\n%s", string(reqDump)))
	}

	// Execute request
	resp, err := t.underlying.RoundTrip(req)
	if err != nil {
		logger.Debug(fmt.Sprintf("Request error: %v", err))
		return resp, err
	}

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Debug(fmt.Sprintf("Error reading response body: %v", err))
		return resp, err
	}

	// Log response
	logger.Debug(fmt.Sprintf("===== RESPONSE (Status %d) =====\n%s\n%s", resp.StatusCode, resp.Proto, string(respBody)))

	// Restore response body so SDK can read it
	resp.Body = io.NopCloser(bytes.NewReader(respBody))

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
