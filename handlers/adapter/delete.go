// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno
//
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/scc-digitalhub/digitalhub-cli-sdk/sdk/config"

	"dhcli/handlers/utils"
	"dhcli/keys"

	crudsvc "github.com/scc-digitalhub/digitalhub-cli-sdk/sdk/services/crud"

	"github.com/spf13/viper"
)

func DeleteHandler(env string, project string, name string, confirm bool, cascade bool, resource string, id string) error {
	endpoint, err := utils.TranslateEndpoint(resource)
	if err != nil {
		return err
	}

	if err := utils.CheckUpdateEnvironment(); err != nil {
		return err
	}
	if err := utils.CheckApiLevel(keys.ApiLevelKey, keys.DeleteMin, keys.DeleteMax); err != nil {
		return err
	}
	if err := utils.CheckCredentials(); err != nil {
		return err
	}

	// Warning per i project senza cascade (comportamento originale)
	if endpoint == "projects" && !cascade {
		log.Println("WARNING: You are deleting a project without the cascade (-c) flag. Resources belonging to the project will not be deleted.")
	}

	// Validazioni
	if endpoint != "projects" && project == "" {
		return errors.New("project is mandatory when performing this operation on resources other than projects")
	}

	confirmationMessage := fmt.Sprintf("Resource %v (%v) will be deleted, proceed? Y/n", id, endpoint)

	delID := id
	delName := name

	// Caso: ID non fornito
	if delID == "" {
		if delName == "" {
			return errors.New("you must specify id or name")
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
		confirmed, err := utils.WaitForConfirmation(confirmationMessage)
		if err != nil {
			return err
		}
		if !confirmed {
			return nil
		}
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
			Resource: endpoint,
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
