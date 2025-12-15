// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"dhcli/core"
	"dhcli/core/facade/adapter"
	"dhcli/core/flags"
	"log"

	"github.com/spf13/cobra"
)

var runCmd = func() *cobra.Command {
	envFlag := flags.NewStringFlag("env", "e", "environment", "")
	projectFlag := flags.NewStringFlag("project", "p", "Mandatory", "")
	fnNameFlag := flags.NewStringFlag("fn-name", "n", "name of the function to run, alternative to Id", "")
	fnIDFlag := flags.NewStringFlag("fn-id", "i", "Id of the function to run, alternative to name", "")
	filePathFlag := flags.NewStringFlag("file", "f", "path to a YAML file containing the resource definition", "")

	cmd := &cobra.Command{
		Use:   "run <task>",
		Short: "Runs a function",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := adapter.RunHandler(
				*envFlag.Value,
				*projectFlag.Value,
				*fnNameFlag.Value,
				*fnIDFlag.Value,
				*filePathFlag.Value,
				args[0],
			)
			if err != nil {
				log.Fatalf("Run failed: %v", err)
			}
		},
	}

	flags.AddFlag(cmd, &envFlag)
	flags.AddFlag(cmd, &projectFlag)
	flags.AddFlag(cmd, &fnNameFlag)
	flags.AddFlag(cmd, &fnIDFlag)
	flags.AddFlag(cmd, &filePathFlag)

	return cmd
}()

func init() {
	core.RegisterCommand(runCmd)
}
