// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	runsvc "github.com/scc-digitalhub/digitalhub-cli-sdk/sdk/services/run"

	"github.com/scc-digitalhub/digitalhub-cli-sdk/sdk/config"

	"dhcli/handlers/utils"

	"github.com/spf13/viper"
)

func LogHandler(env string, project string, container string, follow bool, id string) error {
	endpoint := utils.TranslateEndpoint("run")

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
		HTTPClient: utils.GetDebugHTTPClient(),
	}

	ctx := context.Background()

	// Nuovo RunService (globale) al posto di LogService
	svc, err := runsvc.NewRunService(ctx, cfg)
	if err != nil {
		return err
	}

	// Track the last printed tail of the log to handle circular buffers
	// We'll search for this tail in new logs to find where we left off
	var lastPrintedTail string
	const tailSize = 200 // track last 200 chars to find in new logs

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

		logStr := string(logContents)

		if lastPrintedTail == "" {
			// First iteration - print all and save the tail
			fmt.Print(logStr)
			if len(logStr) > tailSize {
				lastPrintedTail = logStr[len(logStr)-tailSize:]
			} else {
				lastPrintedTail = logStr
			}
		} else {
			// Find where we left off by searching for the tail
			idx := strings.Index(logStr, lastPrintedTail)
			if idx != -1 {
				// Found the tail, print only new content after it
				newContent := logStr[idx+len(lastPrintedTail):]
				fmt.Print(newContent)
			} else {
				// Tail not found - circular buffer wrapped around
				// Print all new logs (might have small duplication but ensures we don't lose logs)
				fmt.Print(logStr)
			}

			// Update tail for next iteration
			if len(logStr) > tailSize {
				lastPrintedTail = logStr[len(logStr)-tailSize:]
			} else {
				lastPrintedTail = logStr
			}
		}

		if !follow {
			return nil
		}

		time.Sleep(5 * time.Second)
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
			Resource: endpoint,
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
				Resource: endpoint,
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
