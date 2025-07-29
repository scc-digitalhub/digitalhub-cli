// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"dhcli/utils"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/viper"
)

func LogHandler(env string, project string, container string, follow bool, resource string, id string) error {
	endpoint := utils.TranslateEndpoint(resource)

	// Load environment and check API level requirements
	utils.CheckUpdateEnvironment()
	utils.CheckApiLevel(utils.ApiLevelKey, utils.LogMin, utils.LogMax)

	if project == "" {
		return errors.New("Project not specified.")
	}

	// Loop requests if following
	for {
		containerLog, err := GetContainerLog(project, endpoint, id, container)

		logContents, err := base64.StdEncoding.DecodeString(containerLog["content"].(string))
		if err != nil {
			return err
		}
		fmt.Printf("%v\n", string(logContents))

		if !follow {
			return nil
		}

		time.Sleep(5 * time.Second)
		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {
			cmd = exec.Command("cmd", "/c", "cls")
		} else {
			cmd = exec.Command("clear")
		}
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

func GetContainerLog(project string, endpoint string, id string, container string) (map[string]interface{}, error) {
	method := "GET"
	url := utils.BuildCoreUrl(project, endpoint, id, nil) + "/logs"
	req := utils.PrepareRequest(method, url, nil, viper.GetString("access_token"))

	body, err := utils.DoRequest(req)
	if err != nil {
		return nil, err
	}
	logs := []interface{}{}
	if err := json.Unmarshal(body, &logs); err != nil {
		return nil, fmt.Errorf("json parsing failed: %w", err)
	}

	// Name of the container to read logs from
	containerName := container
	if containerName == "" {
		// If container is not specified, print main container's logs
		// Get resource to figure out the main container's name
		method := "GET"
		url := utils.BuildCoreUrl(project, endpoint, id, nil)
		req := utils.PrepareRequest(method, url, nil, viper.GetString("access_token"))
		body, err := utils.DoRequest(req)
		if err != nil {
			return nil, err
		}

		var m map[string]interface{}
		if err := json.Unmarshal(body, &m); err != nil {
			return nil, err
		}

		spec := m["spec"].(map[string]interface{})
		task := spec["task"].(string)
		taskFormatted := strings.ReplaceAll(task[:strings.Index(task, ":")], "+", "")

		containerName = fmt.Sprintf("c-%v-%v", taskFormatted, id)
	}

	// Loop over logs to find the correct one
	for _, entry := range logs {
		entryMap := entry.(map[string]interface{})
		status := entryMap["status"]
		statusMap := status.(map[string]interface{})
		entryContainer := statusMap["container"].(string)

		if containerName == entryContainer {
			return entryMap, nil
		}
	}

	return nil, fmt.Errorf("Container not found.")
}
