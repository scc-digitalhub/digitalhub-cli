// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"dhcli/utils"
	"github.com/spf13/viper"
)

func StopHandler(env string, project string, resource string, id string) error {
	endpoint := utils.TranslateEndpoint(resource)

	// Load environment and check API level requirements
	utils.CheckUpdateEnvironment()
	utils.CheckApiLevel(utils.ApiLevelKey, utils.StopMin, utils.StopMax)

	// Request
	method := "POST"
	url := utils.BuildCoreUrl(project, endpoint, id, nil) + "/stop"
	req := utils.PrepareRequest(method, url, nil, viper.GetString("access_token"))

	resp, err := utils.DoRequest(req)
	if err != nil {
		return err
	}

	// Parse response to check new state
	return utils.PrintResponseState(resp)
}
