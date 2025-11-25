// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"context"
	"dhcli/sdk/config"
	sdk "dhcli/sdk/services/run"
	"dhcli/utils"
	"encoding/json"
	"errors"
	"log"
	"os"

	"github.com/spf13/viper"
	"sigs.k8s.io/yaml"
)

func RunHandler(env string, project string, functionName string, functionId string, filePath string, task string) error {
	endpoint := utils.TranslateEndpoint("run")

	utils.CheckUpdateEnvironment()
	utils.CheckApiLevel(utils.ApiLevelKey, utils.CreateMin, utils.CreateMax)

	if project == "" {
		return errors.New("project not specified")
	}

	var inputSpec map[string]interface{}
	if filePath != "" {
		file, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}
		jsonBytes, err := yaml.YAMLToJSON(file)
		if err != nil {
			return err
		}
		var m map[string]interface{}
		if err := json.Unmarshal(jsonBytes, &m); err != nil {
			return err
		}
		if s, ok := m["spec"].(map[string]interface{}); ok {
			inputSpec = s
		} else {
			inputSpec = map[string]interface{}{}
		}
	} else {
		inputSpec = map[string]interface{}{}
	}

	cfg := config.Config{
		Core: config.CoreConfig{
			BaseURL:     viper.GetString(utils.DhCoreEndpoint),
			APIVersion:  viper.GetString(utils.DhCoreApiVersion),
			AccessToken: viper.GetString(utils.DhCoreAccessToken),
		},
	}

	svc, err := sdk.NewRunService(context.Background(), cfg)
	if err != nil {
		return err
	}

	req := sdk.RunRequest{
		Project:              project,
		TaskKind:             task,
		FunctionID:           functionId,
		FunctionName:         functionName,
		InputSpec:            inputSpec,
		ResolvedRunsEndpoint: endpoint,
	}

	if err := svc.Run(context.Background(), req); err != nil {
		return err
	}

	log.Println("Created successfully.")
	return nil
}
