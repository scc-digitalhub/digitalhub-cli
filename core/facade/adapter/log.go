// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"context"
	"dhcli/sdk/config"
	runsvc "dhcli/sdk/services/run"
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

	utils.CheckUpdateEnvironment()
	utils.CheckApiLevel(utils.ApiLevelKey, utils.LogMin, utils.LogMax)

	if project == "" {
		return errors.New("project not specified")
	}

	// Bridge viper -> sdk config
	cfg := config.Config{
		Core: config.CoreConfig{
			BaseURL:     viper.GetString(utils.DhCoreEndpoint),
			APIVersion:  viper.GetString(utils.DhCoreApiVersion),
			AccessToken: viper.GetString(utils.DhCoreAccessToken),
		},
	}

	ctx := context.Background()

	// Nuovo RunService (globale) al posto di LogService
	svc, err := runsvc.NewRunService(ctx, cfg)
	if err != nil {
		return err
	}

	// Loop requests if following
	for {
		containerLog, err := getContainerLogAdapter(ctx, svc, project, endpoint, id, container)
		if err != nil {
			return err
		}

		rawContent, ok := containerLog["content"].(string)
		if !ok {
			return errors.New("invalid log entry: missing or invalid content field")
		}

		logContents, err := base64.StdEncoding.DecodeString(rawContent)
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
		_ = cmd.Run()
	}
}

// same semantics as old GetContainerLog, but using RunService
func getContainerLogAdapter(
	ctx context.Context,
	svc *runsvc.RunService,
	project string,
	endpoint string,
	id string,
	container string,
) (map[string]interface{}, error) {

	// 1) GET /logs
	logBody, _, err := svc.GetLogs(ctx, runsvc.LogRequest{
		RunResourceRequest: runsvc.RunResourceRequest{
			Project:  project,
			Endpoint: endpoint,
			ID:       id,
		},
	})
	if err != nil {
		return nil, err
	}

	var logs []interface{}
	if err := json.Unmarshal(logBody, &logs); err != nil {
		return nil, fmt.Errorf("json parsing failed: %w", err)
	}

	// 2) Determine container name
	containerName := container
	if containerName == "" {
		// se container non specificato, calcola il "main" container come nel vecchio codice

		resBody, _, err := svc.GetResource(ctx, runsvc.LogRequest{
			RunResourceRequest: runsvc.RunResourceRequest{
				Project:  project,
				Endpoint: endpoint,
				ID:       id,
			},
		})
		if err != nil {
			return nil, err
		}

		var m map[string]interface{}
		if err := json.Unmarshal(resBody, &m); err != nil {
			return nil, err
		}

		spec, ok := m["spec"].(map[string]interface{})
		if !ok {
			return nil, errors.New("invalid resource: missing spec")
		}

		task, ok := spec["task"].(string)
		if !ok {
			return nil, errors.New("invalid resource: missing task in spec")
		}

		idx := strings.Index(task, ":")
		if idx == -1 {
			return nil, errors.New("invalid task format in spec")
		}

		taskFormatted := strings.ReplaceAll(task[:idx], "+", "")
		containerName = fmt.Sprintf("c-%v-%v", taskFormatted, id)
	}

	// 3) Loop over logs to find the correct one
	for _, entry := range logs {
		entryMap, ok := entry.(map[string]interface{})
		if !ok {
			continue
		}
		statusVal, ok := entryMap["status"].(map[string]interface{})
		if !ok {
			continue
		}
		entryContainer, _ := statusVal["container"].(string)

		if containerName == entryContainer {
			return entryMap, nil
		}
	}

	return nil, fmt.Errorf("container not found")
}
