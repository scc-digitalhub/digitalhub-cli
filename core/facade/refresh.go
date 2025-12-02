// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package facade

import (
	"dhcli/sdk/utils"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/viper"
)

func RefreshHandler() error {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("client_id", viper.GetString(utils.DhCoreClientId))
	data.Set("refresh_token", viper.GetString(utils.DhCoreRefreshToken))

	resp, err := http.Post(viper.GetString(utils.Oauth2TokenEndpoint), "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
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
	json.Unmarshal(body, &responseJson)
	viper.Set(utils.DhCoreAccessToken, responseJson["access_token"].(string))
	viper.Set(utils.DhCoreRefreshToken, responseJson["refresh_token"].(string))

	err = utils.UpdateIniSectionFromViper(viper.AllKeys())
	if err != nil {
		return err
	}

	log.Printf("Token refreshed.\n")
	return nil
}
