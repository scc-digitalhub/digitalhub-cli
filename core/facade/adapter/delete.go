// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno
//
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/scc-digitalhub/digitalhub-cli-sdk/sdk/config"

	"github.com/scc-digitalhub/digitalhub-cli-sdk/sdk/utils"

	crudsvc "github.com/scc-digitalhub/digitalhub-cli-sdk/sdk/services/crud"

	"github.com/spf13/viper"
)

func DeleteHandler(env string, project string, name string, confirm bool, cascade bool, resource string, id string) error {
	endpoint := utils.TranslateEndpoint(resource)

	utils.CheckUpdateEnvironment()
	utils.CheckApiLevel(utils.ApiLevelKey, utils.DeleteMin, utils.DeleteMax)

	// Warning per i project senza cascade (comportamento originale)
	if endpoint == "projects" && !cascade {
		log.Println("WARNING: You are deleting a project without the cascade (-c) flag. Resources belonging to the project will not be deleted.")
	}

	// Validazioni
	if endpoint != "projects" && project == "" {
		log.Println("Project is mandatory when performing this operation on resources other than projects.")
		os.Exit(1)
	}

	confirmationMessage := fmt.Sprintf("Resource %v (%v) will be deleted, proceed? Y/n", id, endpoint)

	delID := id
	delName := name

	// Caso: ID non fornito
	if delID == "" {
		if delName == "" {
			log.Println("You must specify id or name.")
			os.Exit(1)
		}

		if endpoint != "projects" {
			confirmationMessage = fmt.Sprintf(
				"All versions of resource '%v' (%v) will be deleted, proceed? Y/n",
				delName, endpoint,
			)
		} else {
			confirmationMessage = fmt.Sprintf(
				"Project '%v' will be deleted, proceed? Y/n",
				delName,
			)

			// Comportamento storico: per i projects → ID = name
			delID = delName
			delName = ""
		}
	}

	// Ask for confirmation
	if !confirm {
		utils.WaitForConfirmation(confirmationMessage)
	}

	// Adapter: viper -> sdk.Config
	cfg := config.Config{
		Core: config.CoreConfig{
			BaseURL:     viper.GetString(utils.DhCoreEndpoint),
			APIVersion:  viper.GetString(utils.DhCoreApiVersion),
			AccessToken: viper.GetString(utils.DhCoreAccessToken),
		},
	}

	ctx := context.Background()

	// Inizializza CrudService
	crud, err := crudsvc.NewCrudService(ctx, cfg)
	if err != nil {
		return fmt.Errorf("sdk init failed: %w", err)
	}

	// Request standardizzata
	req := crudsvc.DeleteRequest{
		ResourceRequest: crudsvc.ResourceRequest{
			Project:  project,
			Endpoint: endpoint,
		},
		ID:      delID,
		Name:    delName,
		Cascade: cascade,
	}

	// Esegui delete tramite CrudService
	if err := crud.Delete(ctx, req); err != nil {
		return err
	}

	log.Println("Deleted successfully.")
	return nil
}
