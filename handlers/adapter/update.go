// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/scc-digitalhub/digitalhub-cli-sdk/sdk/config"

	"dhcli/handlers/utils"
	"dhcli/keys"

	crudsvc "github.com/scc-digitalhub/digitalhub-cli-sdk/sdk/services/crud"

	"github.com/spf13/viper"
	"sigs.k8s.io/yaml"
)

func UpdateHandler(env string, project string, filePath string, resource string, id string) error {
	endpoint, err := utils.TranslateEndpoint(resource)
	if err != nil {
		return err
	}

	if err := utils.CheckUpdateEnvironment(); err != nil {
		return err
	}
	if err := utils.CheckApiLevel(keys.ApiLevelKey, keys.UpdateMin, keys.UpdateMax); err != nil {
		return err
	}
	if err := utils.CheckCredentials(); err != nil {
		return err
	}

	if filePath == "" {
		return errors.New("input file not specified")
	}
	if endpoint != "projects" && project == "" {
		return errors.New("project is mandatory when performing this operation on resources other than projects")
	}

	file, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read YAML file: %w", err)
	}
	jsonBytes, err := yaml.YAMLToJSON(file)
	if err != nil {
		return fmt.Errorf("failed to convert YAML to JSON: %w", err)
	}

	var jsonMap map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &jsonMap); err != nil {
		return fmt.Errorf("failed to parse after JSON conversion: %w", err)
	}

	if jsonMap["id"] != nil && jsonMap["id"] != id {
		return fmt.Errorf(
			"specified ID (%v) and ID found in file (%v) do not match; are you sure you are trying to update the correct resource?",
			id, jsonMap["id"],
		)
	}

	delete(jsonMap, "user")
	if endpoint != "projects" {
		jsonMap["project"] = project
	}

	body, err := json.Marshal(jsonMap)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	// Bridge Viper → sdk.Config (retrocomp.)
	cfg := config.Config{
		Core: config.CoreConfig{
			BaseURL:     viper.GetString(keys.DhCoreEndpoint),
			APIVersion:  viper.GetString(keys.DhCoreApiVersion),
			AccessToken: viper.GetString(keys.DhCoreAccessToken),
		},
		HTTPClient: utils.GetDebugHTTPClient(),
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
			Resource: endpoint,
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
