// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"context"
	"log"
)

func UpdateHandler(env string, project string, filePath string, resource string, id string) error {
	return UpdateWithOptionsHandler(UpdateOptions{
		Env:      env,
		Project:  project,
		FilePath: filePath,
		Resource: resource,
		ID:       id,
	})
}

func UpdateWithOptionsHandler(opts UpdateOptions) error {
	respBody, err := UpdateResource(context.Background(), opts)
	if err != nil {
		return err
	}
	log.Println("Updated successfully.")
	return PrintMutationOutput(respBody, opts.Output, "Updated successfully.")
}
