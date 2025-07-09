// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"dhcli/utils"
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
)

func RunLogsHandler(env string, project string, id string) error {
	// Check that CLI has permission to handle runs
	endpoint := utils.TranslateEndpoint("runs")

	// Load environment and check API level requirements
	utils.CheckUpdateEnvironment()
	utils.CheckApiLevel(utils.ApiLevelKey, utils.RunLogsMin, utils.RunLogsMax)

	// Request
	method := "GET"
	url := utils.BuildCoreUrl(project, endpoint, id, nil) + "/logs"
	req := utils.PrepareRequest(method, url, nil, viper.GetString("access_token"))

	body, err := utils.DoRequest(req)
	if err != nil {
		return err
	}
	logs := []interface{}{}
	if err := json.Unmarshal(body, &logs); err != nil {
		return fmt.Errorf("json parsing failed: %w", err)
	}

	printJSONList(logs)

	return nil
}
