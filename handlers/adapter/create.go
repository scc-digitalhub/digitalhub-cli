// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"context"
	"log"
)

func CreateHandler(env string, project string, name string, filePath string, resetID bool, resource string) error {
	return CreateWithOptionsHandler(CreateOptions{
		Env:      env,
		Project:  project,
		Name:     name,
		FilePath: filePath,
		ResetID:  resetID,
		Resource: resource,
	})
}

func CreateWithOptionsHandler(opts CreateOptions) error {
	respBody, err := CreateResource(context.Background(), opts)
	if err != nil {
		return err
	}
	log.Println("Created successfully.")
	return PrintMutationOutput(respBody, opts.Output, "Created successfully.")
}
