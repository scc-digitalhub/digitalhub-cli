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
	"os"
	"strings"
	"time"

	runsvc "github.com/scc-digitalhub/digitalhub-cli-sdk/sdk/services/run"

	"github.com/scc-digitalhub/digitalhub-cli-sdk/sdk/config"

	"dhcli/handlers/utils"
	"dhcli/keys"

	"github.com/spf13/viper"
)

func LogHandler(env string, project string, container string, follow bool, id string) error {
	endpoint := utils.TranslateEndpoint("run")

	utils.CheckUpdateEnvironment()
	utils.CheckApiLevel(keys.ApiLevelKey, keys.LogMin, keys.LogMax)
	if err := utils.CheckCredentials(); err != nil {
		return err
	}

	if project == "" {
		return errors.New("project not specified")
	}

	// Bridge viper -> sdk config
	cfg := config.Config{
		Core: config.CoreConfig{
			BaseURL:     viper.GetString(keys.DhCoreEndpoint),
			APIVersion:  viper.GetString(keys.DhCoreApiVersion),
			AccessToken: viper.GetString(keys.DhCoreAccessToken),
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

	// 2) Collect all log entries — the top-level "id" field is the container name
	type entryWithContainer struct {
		name  string
		entry map[string]interface{}
	}
	var entries []entryWithContainer
	for _, raw := range logs {
		entryMap, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := entryMap["id"].(string)
		if name == "" {
			continue
		}
		entries = append(entries, entryWithContainer{name: name, entry: entryMap})
	}

	if len(entries) == 0 {
		return nil, fmt.Errorf("no log entries with a container found")
	}

	// 3) Pick container
	if container == "" {
		if len(entries) > 1 {
			names := make([]string, len(entries))
			for i, e := range entries {
				names[i] = e.name
			}
			fmt.Fprintf(os.Stderr, "More than one container found: %s, picking %s...\n",
				strings.Join(names, ", "), entries[0].name)
		}
		return entries[0].entry, nil
	}

	for _, e := range entries {
		if e.name == container {
			return e.entry, nil
		}
	}

	return nil, fmt.Errorf("container %q not found", container)
}
