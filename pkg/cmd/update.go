// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"dhcli/handlers/adapter"
	"dhcli/pkg"
	"dhcli/pkg/flags"
	"log"

	"dhcli/handlers/utils"

	"github.com/spf13/cobra"
)

var updateCmd = func() *cobra.Command {
	envFlag := flags.NewStringFlag("env", "e", "environment", "")
	projectFlag := flags.NewStringFlag("project", "p", "Mandatory for resources other than projects", "")
	fileFlag := flags.NewStringFlag("file", "f", "path to the YAML file containing the resource data to be updated; mandatory", "")

	cmd := &cobra.Command{
		Use:   "update <resource> <id>",
		Short: "Update a specific resource using data from a YAML file",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			project := utils.ResolveProject(*projectFlag.Value)
			err := adapter.UpdateHandler(
				*envFlag.Value,
				project,
				*fileFlag.Value,
				args[0],
				args[1],
			)
			if err != nil {
				log.Fatalf("Update failed: %v", err)
			}
		},
	}

	flags.AddFlag(cmd, &envFlag)
	flags.AddFlag(cmd, &projectFlag)
	flags.AddFlag(cmd, &fileFlag)

	return cmd
}()

func init() {
	pkg.RegisterCommand(updateCmd)
}
