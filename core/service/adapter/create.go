// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"context"
	"dhcli/sdk"
	"dhcli/utils"
	"fmt"
	"log"
	"os"

	"github.com/spf13/viper"
)

func CreateHandler(env string, project string, name string, filePath string, resetId bool, resource string) error {
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
	cfg := sdk.Config{
		Core: sdk.CoreConfig{
			BaseURL:     viper.GetString(utils.DhCoreEndpoint),
			APIVersion:  viper.GetString(utils.DhCoreApiVersion),
			AccessToken: viper.GetString(utils.DhCoreAccessToken),
		},
	}

	svc, err := sdk.NewCreateService(context.Background(), cfg)
	if err != nil {
		return fmt.Errorf("sdk init failed: %w", err)
	}

	req := sdk.CreateRequest{
		Project:  project,
		Endpoint: endpoint,
		Name:     name,
		FilePath: filePath,
		ResetID:  resetId,
	}

	if err := svc.Create(req); err != nil {
		return err
	}

	log.Println("Created successfully.")
	return nil
}
