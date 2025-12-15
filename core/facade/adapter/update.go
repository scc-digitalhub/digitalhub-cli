// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/scc-digitalhub/digitalhub-cli-sdk/sdk/config"

	"github.com/scc-digitalhub/digitalhub-cli-sdk/sdk/utils"

	crudsvc "github.com/scc-digitalhub/digitalhub-cli-sdk/sdk/services/crud"

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
		log.Printf(
			"Error: specified ID (%v) and ID found in file (%v) do not match. "+
				"Are you sure you are trying to update the correct resource?\n",
			id, jsonMap["id"],
		)
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
	cfg := config.Config{
		Core: config.CoreConfig{
			BaseURL:     viper.GetString(utils.DhCoreEndpoint),
			APIVersion:  viper.GetString(utils.DhCoreApiVersion),
			AccessToken: viper.GetString(utils.DhCoreAccessToken),
		},
	}

	ctx := context.Background()

	// Usa il CrudService al posto del vecchio UpdateService
	crud, err := crudsvc.NewCrudService(ctx, cfg)
	if err != nil {
		return err
	}

	req := crudsvc.UpdateRequest{
		ResourceRequest: crudsvc.ResourceRequest{
			Project:  project,
			Endpoint: endpoint,
		},
		ID:   id,
		Body: body,
	}

	if err := crud.Update(ctx, req); err != nil {
		return err
	}

	log.Println("Updated successfully.")
	return nil
}
