// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"context"
	"log"
)

func RunHandler(env string, project string, functionName string, functionId string, filePath string, task string) error {
	return RunWithOptionsHandler(RunOptions{
		Env:      env,
		Project:  project,
		Name:     functionName,
		ID:       functionId,
		FilePath: filePath,
		Task:     task,
	})
}

func RunWithOptionsHandler(opts RunOptions) error {
	respBody, err := ExecuteRun(context.Background(), opts)
	if err != nil {
		return err
	}
	log.Println("Created successfully.")
	return PrintMutationOutput(respBody, opts.Output, "Created successfully.")
}
