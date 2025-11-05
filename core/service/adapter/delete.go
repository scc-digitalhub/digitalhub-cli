// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno

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

func DeleteHandler(env string, project string, name string, confirm bool, cascade bool, resource string, id string) error {
	endpoint := utils.TranslateEndpoint(resource)

	utils.CheckUpdateEnvironment()
	utils.CheckApiLevel(utils.ApiLevelKey, utils.DeleteMin, utils.DeleteMax)

	if endpoint == "projects" && !cascade {
		log.Println("WARNING: You are deleting a project without the cascade (-c) flag. Resources belonging to the project will not be deleted.")
	}

	if endpoint != "projects" && project == "" {
		log.Println("Project is mandatory when performing this operation on resources other than projects.")
		os.Exit(1)
	}

	confirmationMessage := fmt.Sprintf("Resource %v (%v) will be deleted, proceed? Y/n", id, endpoint)

	delID := id
	delName := name

	if delID == "" {
		if delName == "" {
			log.Println("You must specify id or name.")
			os.Exit(1)
		}
		if endpoint != "projects" {
			confirmationMessage = fmt.Sprintf("All versions of resource '%v' (%v) will be deleted, proceed? Y/n", delName, endpoint)
		} else {
			confirmationMessage = fmt.Sprintf("Project '%v' will be deleted, proceed? Y/n", delName)
			// per projects, il backend vuole l'id = name (comportamento precedente)
			delID = delName
			delName = ""
		}
	}

	if !confirm {
		utils.WaitForConfirmation(confirmationMessage)
	}

	// Mappa Viper → sdk.Config (retrocompat con INI/env/flags già caricati)
	cfg := sdk.Config{
		Core: sdk.CoreConfig{
			BaseURL:     viper.GetString(utils.DhCoreEndpoint),
			APIVersion:  viper.GetString(utils.DhCoreApiVersion),
			AccessToken: viper.GetString(utils.DhCoreAccessToken),
		},
	}

	svc, err := sdk.NewDeleteService(context.Background(), cfg)
	if err != nil {
		return fmt.Errorf("sdk init failed: %w", err)
	}

	req := sdk.DeleteRequest{
		Project:  project,
		Endpoint: endpoint,
		ID:       delID,
		Name:     delName,
		Cascade:  cascade,
	}

	if err := svc.Delete(context.Background(), req); err != nil {
		return err
	}

	log.Println("Deleted successfully.")
	return nil
}
