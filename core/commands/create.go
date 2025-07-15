// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"dhcli/core"
	"dhcli/core/flags"
	"dhcli/core/service"
	"log"

	"github.com/spf13/cobra"
)

var createCmd = func() *cobra.Command {

	// Declare cmd all flags

	envFlag := flags.NewStringFlag("env", "e", "environment", "")
	projectFlag := flags.NewStringFlag("project", "p", "project", "")
	nameFlag := flags.NewStringFlag("name", "n", "name", "")
	resetIdFlag := flags.NewBoolFlag("reset-id", "r", "if set, removes the id field from the file to ensure the server assigns a new one", false)
	fileFlag := flags.NewStringFlag("file", "f", "path to a YAML file containing the resource definition", "")

	cmd := &cobra.Command{
		Use:   "create <resource>",
		Short: "Creates a new resource from a YAML file (or a name for projects)",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := service.CreateHandler(
				*envFlag.Value,
				*projectFlag.Value,
				*nameFlag.Value,
				*fileFlag.Value,
				*resetIdFlag.Value,
				args[0],
			)
			if err != nil {
				log.Fatalf("Create failed: %v", err)
			}
		},
	}

	// Add all flags
	flags.AddFlag(cmd, &envFlag)
	flags.AddFlag(cmd, &projectFlag)
	flags.AddFlag(cmd, &nameFlag)
	flags.AddFlag(cmd, &resetIdFlag)
	flags.AddFlag(cmd, &fileFlag)

	return cmd
}()

func init() {
	core.RegisterCommand(createCmd)
}
