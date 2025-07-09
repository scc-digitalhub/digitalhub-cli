// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"dhcli/utils"
)

func StopHandler(env string, project string, resource string, id string) error {
	endpoint := utils.TranslateEndpoint(resource)

	// Load environment and check API level requirements
	cfg, section := utils.LoadIniConfig([]string{env})
	utils.CheckUpdateEnvironment(cfg, section)
	utils.CheckApiLevel(section, utils.StopMin, utils.StopMax)

	// Request
	method := "POST"
	url := utils.BuildCoreUrl(section, project, endpoint, id, nil) + "/stop"
	req := utils.PrepareRequest(method, url, nil, section.Key("access_token").String())

	resp, err := utils.DoRequest(req)
	if err != nil {
		return err
	}

	// Parse response to check new state
	return utils.PrintResponseState(resp)
}
