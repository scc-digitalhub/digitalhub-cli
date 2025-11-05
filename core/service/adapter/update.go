// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler

// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"context"
	"dhcli/sdk"
	"dhcli/utils"
	"encoding/json"
	"log"
	"os"

	"github.com/spf13/viper"
	"sigs.k8s.io/yaml"
)

func UpdateHandler(env string, project string, filePath string, resource string, id string) error {
	endpoint := utils.TranslateEndpoint(resource)

	utils.CheckUpdateEnvironment()
	utils.CheckApiLevel(utils.ApiLevelKey, utils.UpdateMin, utils.UpdateMax)

	if filePath == "" {
		log.Println("Input file not specified.")
		os.Exit(1)
	}
	if endpoint != "projects" && project == "" {
		log.Println("Project is mandatory when performing this operation on resources other than projects.")
		os.Exit(1)
	}

	file, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Failed to read YAML file: %v\n", err)
		os.Exit(1)
	}
	jsonBytes, err := yaml.YAMLToJSON(file)
	if err != nil {
		log.Printf("Failed to convert YAML to JSON: %v\n", err)
		os.Exit(1)
	}

	var jsonMap map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &jsonMap); err != nil {
		log.Printf("Failed to parse after JSON conversion: %v\n", err)
		os.Exit(1)
	}

	if jsonMap["id"] != nil && jsonMap["id"] != id {
		log.Printf("Error: specified ID (%v) and ID found in file (%v) do not match. Are you sure you are trying to update the correct resource?\n", id, jsonMap["id"])
		os.Exit(1)
	}

	delete(jsonMap, "user")
	if endpoint != "projects" {
		jsonMap["project"] = project
	}

	body, err := json.Marshal(jsonMap)
	if err != nil {
		log.Printf("Failed to marshal: %v\n", err)
		os.Exit(1)
	}

	// Bridge Viper → sdk.Config (retrocomp.)
	cfg := sdk.Config{
		Core: sdk.CoreConfig{
			BaseURL:     viper.GetString(utils.DhCoreEndpoint),
			APIVersion:  viper.GetString(utils.DhCoreApiVersion),
			AccessToken: viper.GetString(utils.DhCoreAccessToken),
		},
	}

	svc, err := sdk.NewUpdateService(context.Background(), cfg)
	if err != nil {
		return err
	}

	req := sdk.UpdateRequest{
		Project:  project,
		Endpoint: endpoint,
		ID:       id,
		Body:     body,
	}

	if err := svc.Update(context.Background(), req); err != nil {
		return err
	}

	log.Println("Updated successfully.")
	return nil
}
