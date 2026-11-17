// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"

	"github.com/scc-digitalhub/digitalhub-cli-sdk/sdk/config"

	"dhcli/handlers/utils"
	"dhcli/keys"

	crudsvc "github.com/scc-digitalhub/digitalhub-cli-sdk/sdk/services/crud"
)

func DeleteHandler(env string, project string, name string, confirm bool, cascade bool, resource string, id string) error {
	return DeleteResource(context.Background(), env, project, name, confirm, cascade, resource, id)
}

func DeleteResource(ctx context.Context, env string, project string, name string, confirm bool, cascade bool, resource string, id string) error {
	endpoint, err := utils.ResolveEndpoint(resource)
	if err != nil {
		return err
	}

	if err := prepareEnvironment(ctx, env, keys.DeleteMin, keys.DeleteMax); err != nil {
		return err
	}

	if endpoint == "projects" && !cascade {
		log.Println("WARNING: You are deleting a project without the cascade (-c) flag. Resources belonging to the project will not be deleted.")
	}

	if endpoint != "projects" && project == "" {
		return errors.New("project is mandatory when performing this operation on resources other than projects")
	}

	confirmationMessage := fmt.Sprintf("Resource %v (%v) will be deleted, proceed? Y/n", id, endpoint)

	delID := id
	delName := name

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
			delID = delName
			delName = ""
		}
	}

	if !confirm {
		utils.WaitForConfirmation(confirmationMessage)
	}

	cfg := NewCoreConfig()
	crud, err := crudsvc.NewCrudService(ctx, config.Config{
		Core:       cfg.Core,
		HTTPClient: cfg.HTTPClient,
	})
	if err != nil {
		return fmt.Errorf("sdk init failed: %w", err)
	}

	req := crudsvc.DeleteRequest{
		ResourceRequest: crudsvc.ResourceRequest{
			Project:  project,
			Resource: endpoint,
		},
		ID:      delID,
		Name:    url.QueryEscape(delName),
		Cascade: cascade,
	}

	if err := crud.Delete(ctx, req); err != nil {
		return err
	}

	log.Println("Deleted successfully.")
	return nil
}
