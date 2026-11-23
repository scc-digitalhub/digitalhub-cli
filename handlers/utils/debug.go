// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// DebugTransport wraps http.RoundTripper to log requests and responses
type DebugTransport struct {
	underlying http.RoundTripper
	trace      bool
}

// RoundTrip implements http.RoundTripper interface
func (t *DebugTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.trace {
		return t.roundTripTrace(req)
	}
	return t.roundTripDebug(req)
}

func (t *DebugTransport) roundTripDebug(req *http.Request) (*http.Response, error) {
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

func (t *DebugTransport) roundTripTrace(req *http.Request) (*http.Response, error) {
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
	_ = resp.Body.Close()
	if err != nil {
		logger.Debug(fmt.Sprintf("Error reading response body: %v", err))
		return nil, err
	}

	// Log response
	logger.Debug(fmt.Sprintf("===== RESPONSE (Status %d) =====\n%s\n%s", resp.StatusCode, resp.Proto, string(respBody)))

	// Restore response body so SDK can read it
	resp.Body = io.NopCloser(bytes.NewReader(respBody))

	return resp, nil
}

// CreateDebugHTTPClient creates an HTTP client with redacted request/response logging
func CreateDebugHTTPClient() *http.Client {
	return &http.Client{
		Transport: &DebugTransport{
			underlying: http.DefaultTransport,
		},
	}
}

// CreateTraceHTTPClient creates an HTTP client with full request/response logging
func CreateTraceHTTPClient() *http.Client {
	return &http.Client{
		Transport: &DebugTransport{
			underlying: http.DefaultTransport,
			trace:      true,
		},
	}
}

// Global debug HTTP client
var debugHTTPClient *http.Client

// EnableDebugMode enables redacted HTTP debug logging and returns the debug client
func EnableDebugMode() *http.Client {
	if debugHTTPClient == nil {
		debugHTTPClient = CreateDebugHTTPClient()
	}
	return debugHTTPClient
}

// EnableTraceMode enables full HTTP trace logging and returns the trace client
func EnableTraceMode() *http.Client {
	debugHTTPClient = CreateTraceHTTPClient()
	return debugHTTPClient
}

// GetDebugHTTPClient returns the global debug HTTP client if debug or trace mode is enabled
func GetDebugHTTPClient() *http.Client {
	return debugHTTPClient
}
