// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"dhcli/utils"
)

func ResumeHandler(env string, project string, resource string, id string) error {
	endpoint := utils.TranslateEndpoint(resource)

	// Load environment and check API level requirements
	cfg, section := utils.LoadIniConfig([]string{env})
	utils.CheckUpdateEnvironment(cfg, section)
	utils.CheckApiLevel(section, utils.ResumeMin, utils.ResumeMax)

	// Request
	method := "POST"
	url := utils.BuildCoreUrl(section, project, endpoint, id, nil) + "/resume"
	req := utils.PrepareRequest(method, url, nil, section.Key("access_token").String())

	_, err := utils.DoRequest(req)
	if err != nil {
		return err
	}

	resp, err := utils.DoRequest(req)
	if err != nil {
		return err
	}

	// Parse response to check new state
	return utils.PrintResponseState(resp)
}
