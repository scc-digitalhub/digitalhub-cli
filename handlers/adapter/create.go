// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"context"
	"errors"
	"log"

	"dhcli/handlers/utils"
	"dhcli/keys"

	"github.com/scc-digitalhub/digitalhub-cli-sdk/sdk/config"

	crudsvc "github.com/scc-digitalhub/digitalhub-cli-sdk/sdk/services/crud"

	"github.com/spf13/viper"
)

func CreateHandler(env string, project string, name string, filePath string, resetID bool, resource string) error {
	endpoint, err := utils.TranslateEndpoint(resource)
	if err != nil {
		return err
	}

	if err := utils.CheckUpdateEnvironment(); err != nil {
		return err
	}
	if err := utils.CheckApiLevel(keys.ApiLevelKey, keys.CreateMin, keys.CreateMax); err != nil {
		return err
	}
	if err := utils.CheckCredentials(); err != nil {
		return err
	}

	if endpoint != "projects" {
		if project == "" {
			return errors.New("project is mandatory when performing this operation on resources other than projects")
		}
		if filePath == "" {
			return errors.New("input file not specified")
		}
	} else if filePath == "" && name == "" {
		return errors.New("must provide either an input file or a name when creating a project")
	}

	// Adapter: viper -> sdk.Config
	cfg := config.Config{
		Core: config.CoreConfig{
			BaseURL:     viper.GetString(keys.DhCoreEndpoint),
			APIVersion:  viper.GetString(keys.DhCoreApiVersion),
			AccessToken: viper.GetString(keys.DhCoreAccessToken),
		},
		HTTPClient: utils.GetDebugHTTPClient(),
	}

	// ctx per il CrudService e per la Create
	ctx := context.Background()

	svc, err := crudsvc.NewCrudService(ctx, cfg)
	if err != nil {
		return err
	}

	req := crudsvc.CreateRequest{
		ResourceRequest: crudsvc.ResourceRequest{
			Project:  project,
			Resource: endpoint,
		},
		Name:     name,
		FilePath: filePath,
		ResetID:  resetID,
	}

	if err := svc.Create(ctx, req); err != nil {
		return err
	}

	log.Println("Created successfully.")
	return nil
}
