// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"bytes"
	"dhcli/utils"
	"encoding/json"
	"errors"
	"fmt"
)

func MetricsHandler(env string, project string, container string, resource string, id string) error {
	endpoint := utils.TranslateEndpoint(resource)

	// Load environment and check API level requirements
	utils.CheckUpdateEnvironment()
	utils.CheckApiLevel(utils.ApiLevelKey, utils.MetricsMin, utils.MetricsMax)

	if project == "" {
		return errors.New("Project not specified.")
	}

	containerLog, err := GetContainerLog(project, endpoint, id, container)
	if err != nil {
		return err
	}

	statusMap := containerLog["status"].(map[string]interface{})
	jsonData, err := json.Marshal(statusMap["metrics"].([]interface{}))

	var pretty bytes.Buffer
	if err := json.Indent(&pretty, jsonData, "", "    "); err != nil {
		return err
	}
	fmt.Println(pretty.String())

	return nil

	/* This calls the /metrics API
	// Request
	method := "GET"
	url := utils.BuildCoreUrl(project, endpoint, id, nil) + "/metrics"
	req := utils.PrepareRequest(method, url, nil, viper.GetString("access_token"))

	body, err := utils.DoRequest(req)
	if err != nil {
		return err
	}

	var pretty bytes.Buffer
	if err := json.Indent(&pretty, body, "", "    "); err != nil {
		return err
	}
	fmt.Println(pretty.String())

	return nil
	*/
}
