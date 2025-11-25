// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"context"
	"dhcli/sdk/config"
	crudsvc "dhcli/sdk/services/crud"
	"dhcli/utils"
	"log"
	"os"

	"github.com/spf13/viper"
)

func CreateHandler(env string, project string, name string, filePath string, resetID bool, resource string) error {
	endpoint := utils.TranslateEndpoint(resource)

	utils.CheckUpdateEnvironment()
	utils.CheckApiLevel(utils.ApiLevelKey, utils.CreateMin, utils.CreateMax)

	if endpoint != "projects" {
		if project == "" {
			log.Println("Project is mandatory when performing this operation on resources other than projects.")
			os.Exit(1)
		}
		if filePath == "" {
			log.Println("Input file not specified.")
			os.Exit(1)
		}
	} else if filePath == "" && name == "" {
		log.Println("Must provide either an input file or a name when creating a project.")
		os.Exit(1)
	}

	// Adapter: viper -> sdk.Config
	cfg := config.Config{
		Core: config.CoreConfig{
			BaseURL:     viper.GetString(utils.DhCoreEndpoint),
			APIVersion:  viper.GetString(utils.DhCoreApiVersion),
			AccessToken: viper.GetString(utils.DhCoreAccessToken),
		},
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
			Endpoint: endpoint,
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
