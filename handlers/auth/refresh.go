// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"dhcli/handlers/utils"
	"dhcli/keys"

	"github.com/spf13/viper"
)

func RefreshHandler() error {
	if viper.GetString(keys.DhCoreRefreshToken) == "" {
		return fmt.Errorf("no refresh token available – please log in first")
	}

	// Read and normalize scopes from config
	raw := viper.GetString(keys.OAuth2ScopesSupported)
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

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("client_id", viper.GetString(keys.DhCoreClientId))
	data.Set("refresh_token", viper.GetString(keys.DhCoreRefreshToken))
	if len(scopes) > 0 {
		data.Set("scope", strings.Join(scopes, " "))
	}

	// Use debug HTTP client if available, otherwise use default
	client := utils.GetDebugHTTPClient()
	if client == nil {
		client = &http.Client{}
	}

	resp, err := client.Post(viper.GetString(keys.OAuth2TokenEndpoint), "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token server error: %s %s", resp.Status, string(body))
	}

	credKeys, err := utils.ApplyTokenResponse(body)
	if err != nil {
		return fmt.Errorf("failed to apply token response: %w", err)
	}
	credKeys = append(credKeys, keys.CredentialsList)
	if err := utils.PersistCurrentEnv(credKeys); err != nil {
		return err
	}

	log.Printf("Token refreshed.\n")
	return nil
}
