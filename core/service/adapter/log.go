// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"context"
	"dhcli/sdk"
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

	cfg := sdk.Config{
		Core: sdk.CoreConfig{
			BaseURL:     viper.GetString(utils.DhCoreEndpoint),
			APIVersion:  viper.GetString(utils.DhCoreApiVersion),
			AccessToken: viper.GetString(utils.DhCoreAccessToken),
		},
	}

	svc, err := sdk.NewLogService(context.Background(), cfg)
	if err != nil {
		return err
	}

	// Loop requests if following
	for {
		containerLog, err := getContainerLogAdapter(context.Background(), svc, project, endpoint, id, container)
		if err != nil {
			return err
		}

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
		_ = cmd.Run()
	}
}

// same semantics as old GetContainerLog, but using sdk.LogService
func getContainerLogAdapter(
	ctx context.Context,
	svc *sdk.LogService,
	project string,
	endpoint string,
	id string,
	container string,
) (map[string]interface{}, error) {

	// GET /logs
	logBody, _, err := svc.GetLogs(ctx, sdk.LogRequest{
		Project:  project,
		Endpoint: endpoint,
		ID:       id,
	})
	if err != nil {
		return nil, err
	}

	logs := []interface{}{}
	if err := json.Unmarshal(logBody, &logs); err != nil {
		return nil, fmt.Errorf("json parsing failed: %w", err)
	}

	// Determine container name
	containerName := container
	if containerName == "" {
		// If container is not specified, compute "main" container as in old code.

		resBody, _, err := svc.GetResource(ctx, sdk.LogRequest{
			Project:  project,
			Endpoint: endpoint,
			ID:       id,
		})
		if err != nil {
			return nil, err
		}

		var m map[string]interface{}
		if err := json.Unmarshal(resBody, &m); err != nil {
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

	return nil, fmt.Errorf("container not found")
}
