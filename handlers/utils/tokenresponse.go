// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"dhcli/keys"

	"github.com/spf13/viper"
)

// Oauth2TokenResponse holds the known standard OAuth2 token response fields.
// Fields declared here are consumed by the struct and NOT passed through to
// dynamic credential storage. Everything else in the response is pass-through.
type Oauth2TokenResponse struct {
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	ClientId     string `json:"client_id"`
	Issuer       string `json:"issuer"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IdToken      string `json:"id_token"`
}

// oauth2KnownFields is the set of JSON field names declared in Oauth2TokenResponse.
// Built once at startup via reflection; used to filter the pass-through map.
var oauth2KnownFields map[string]bool

func init() {
	oauth2KnownFields = make(map[string]bool)
	rt := reflect.TypeOf(Oauth2TokenResponse{})
	for i := 0; i < rt.NumField(); i++ {
		if tag := rt.Field(i).Tag.Get("json"); tag != "" && tag != "-" {
			oauth2KnownFields[tag] = true
		}
	}
}

// ApplyTokenResponse processes a raw OAuth2 token response body:
//   - access_token, refresh_token, id_token are stored with the dhcore_ prefix
//   - token_type, expires_in, client_id, issuer are discarded (consumed by struct)
//   - all remaining fields are stored in Viper as-is (dynamic backend credentials)
//
// credentials_list in Viper is updated to the union of its current value and the
// new credential keys, so re-login never shrinks the recorded credential set.
//
// Returns the merged credentials_list as a slice (suitable for PersistCurrentEnv
// additionalKeys so that credentials_list itself is written to the INI).
func ApplyTokenResponse(body []byte) ([]string, error) {
	var oauth Oauth2TokenResponse
	if err := json.Unmarshal(body, &oauth); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	var full map[string]interface{}
	if err := json.Unmarshal(body, &full); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	// Remove fields handled by the struct from the pass-through map.
	for k := range oauth2KnownFields {
		delete(full, k)
	}

	var credKeys []string

	// Compute expires_at from expires_in if present.
	if oauth.ExpiresIn > 0 {
		expiresAt := time.Now().Add(time.Duration(oauth.ExpiresIn) * time.Second).UTC().Format(time.RFC3339)
		viper.Set(keys.DhCoreExpiresAt, expiresAt)
		credKeys = append(credKeys, keys.DhCoreExpiresAt)
	}

	// Standard tokens: stored with dhcore_ prefix.
	if oauth.AccessToken != "" {
		viper.Set(keys.DhCoreAccessToken, oauth.AccessToken)
		credKeys = append(credKeys, keys.DhCoreAccessToken)
	}
	if oauth.RefreshToken != "" {
		viper.Set(keys.DhCoreRefreshToken, oauth.RefreshToken)
		credKeys = append(credKeys, keys.DhCoreRefreshToken)
	}
	if oauth.IdToken != "" {
		viper.Set(keys.DhCoreIdToken, oauth.IdToken)
		credKeys = append(credKeys, keys.DhCoreIdToken)
	}

	// Pass-through: all remaining fields stored as-is (dynamic backend credentials).
	for k, v := range full {
		viper.Set(k, fmt.Sprint(v))
		credKeys = append(credKeys, k)
	}

	// Update credentials_list: union with existing so re-login never shrinks it.
	existing := SplitCSV(viper.GetString(keys.CredentialsList))
	merged := unionStringSlice(existing, credKeys)
	viper.Set(keys.CredentialsList, strings.Join(merged, ","))

	return merged, nil
}

// SplitCSV splits a comma-separated string into a trimmed, non-empty slice.
func SplitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// unionStringSlice returns a new slice containing all elements of a and b,
// without duplicates, preserving order (a first, then new elements from b).
func unionStringSlice(a, b []string) []string {
	seen := make(map[string]bool, len(a)+len(b))
	out := make([]string, 0, len(a)+len(b))
	for _, v := range a {
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	for _, v := range b {
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}
