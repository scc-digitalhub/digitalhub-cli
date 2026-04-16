// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"dhcli/handlers/utils"

	"github.com/spf13/viper"
)

func RefreshHandler() error {
	// Read and normalize scopes from config
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

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("client_id", viper.GetString(utils.DhCoreClientId))
	data.Set("refresh_token", viper.GetString(utils.DhCoreRefreshToken))
	if len(scopes) > 0 {
		data.Set("scope", strings.Join(scopes, " "))
	}

	// Use debug HTTP client if available, otherwise use default
	client := utils.GetDebugHTTPClient()
	if client == nil {
		client = &http.Client{}
	}

	resp, err := client.Post(viper.GetString(utils.Oauth2TokenEndpoint), "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
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

	var responseJson map[string]interface{}
	if err := json.Unmarshal(body, &responseJson); err != nil {
		return fmt.Errorf("json parse error: %w", err)
	}

	// Map all token response fields into Viper (not just access_token and refresh_token)
	for k, v := range responseJson {
		key := k
		if mapped, ok := utils.DhCoreMap[k]; ok {
			key = mapped
		}
		viper.Set(key, fmt.Sprint(v))
	}

	// Persist all config keys to ini file
	if err := utils.UpdateIniSectionFromViper(viper.AllKeys()); err != nil {
		return err
	}

	log.Printf("Token refreshed.\n")
	return nil
}
