// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// authMode describes how the current environment is authenticated.
type authMode int

const (
	authModePublic authMode = iota // no credentials configured
	authModeOAuth2                 // bearer token (access_token / refresh_token)
	authModeBasic                  // HTTP Basic auth (user + password)
)

// detectAuthMode returns the active authentication mode based on what is
// currently stored in Viper.
func detectAuthMode() authMode {
	if viper.GetString(DhCoreAccessToken) != "" || viper.GetString(DhCoreRefreshToken) != "" {
		return authModeOAuth2
	}
	if viper.GetString(DhCoreUser) != "" && viper.GetString(DhCorePassword) != "" {
		return authModeBasic
	}
	return authModePublic
}

// CheckCredentials probes GET /api/auth to validate the current credentials.
//
// Three auth modes are supported:
//   - OAuth2  (DHCORE_ACCESS_TOKEN or DHCORE_REFRESH_TOKEN present): sends a
//     Bearer header; on 401 attempts a refresh_token grant and persists the new
//     tokens.
//   - Basic   (DHCORE_USER + DHCORE_PASSWORD present): sends an Authorization:
//     Basic header; on 401 returns an error because the credentials are wrong
//     and cannot be auto-renewed.
//   - Public  (nothing configured): sends no Authorization header; on 401
//     returns nil and lets the real operation surface its own error.
//
// Any non-401 response (including network errors) is treated as "proceed" so
// that transient server issues do not block the caller.
func CheckCredentials() error {
	endpoint := viper.GetString(DhCoreEndpoint)
	if endpoint == "" {
		return nil
	}
	logger.Info(fmt.Sprintf("Checking credentials for %v ...", endpoint))

	authURL := strings.TrimRight(endpoint, "/") + "/api/auth"

	client := GetDebugHTTPClient()
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}

	req, err := http.NewRequest(http.MethodGet, authURL, nil)
	if err != nil {
		return nil // misconfigured URL; let the real call fail
	}

	mode := detectAuthMode()
	switch mode {
	case authModeOAuth2:
		if tok := viper.GetString(DhCoreAccessToken); tok != "" {
			req.Header.Set("Authorization", "Bearer "+tok)
		}
	case authModeBasic:
		req.SetBasicAuth(viper.GetString(DhCoreUser), viper.GetString(DhCorePassword))
	case authModePublic:
		// no Authorization header
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil // network error; let the real call fail
	}
	defer resp.Body.Close()
	_, _ = io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusUnauthorized {
		return nil
	}

	switch mode {
	case authModeOAuth2:
		return doRefresh()
	case authModeBasic:
		return fmt.Errorf("authentication failed: invalid username or password")
	default: // authModePublic
		return nil // no credentials to renew; let the real call fail
	}
}

// doRefresh performs a refresh_token grant against the OAuth2 token endpoint
// and persists the resulting tokens to viper and the ini file.
func doRefresh() error {
	logger.Info(fmt.Sprintf("Refreshing credentials for %v ...", viper.GetString(DhCoreEndpoint)))

	refreshToken := viper.GetString(DhCoreRefreshToken)
	if refreshToken == "" {
		return fmt.Errorf("session expired and no refresh token available – please log in again")
	}

	tokenURL := viper.GetString(Oauth2TokenEndpoint)
	if tokenURL == "" {
		return fmt.Errorf("oauth2_token_endpoint not configured – please log in again")
	}

	clientID := viper.GetString(DhCoreClientId)

	raw := viper.GetString("scopes_supported")
	var scopes []string
	if raw != "" {
		for _, s := range strings.FieldsFunc(raw, func(r rune) bool {
			return r == ',' || r == ' ' || r == '\n' || r == '\t'
		}) {
			if s != "" {
				scopes = append(scopes, s)
			}
		}
	}

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("client_id", clientID)
	data.Set("refresh_token", refreshToken)
	if len(scopes) > 0 {
		data.Set("scope", strings.Join(scopes, " "))
	}

	client := GetDebugHTTPClient()
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}

	resp, err := client.Post(tokenURL, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("token refresh request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read refresh response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token refresh failed (%s) – please log in again", resp.Status)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(body, &m); err != nil {
		return fmt.Errorf("failed to parse refresh token response: %w", err)
	}

	for k, v := range m {
		key := k
		if mapped, ok := DhCoreMap[k]; ok {
			key = mapped
		}
		viper.Set(key, fmt.Sprint(v))
	}

	if err := UpdateIniSectionFromViper(viper.AllKeys()); err != nil {
		return fmt.Errorf("failed to persist refreshed tokens: %w", err)
	}

	return nil
}
