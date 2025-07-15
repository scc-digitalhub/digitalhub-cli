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

var updateCmd = func() *cobra.Command {
	envFlag := flags.NewStringFlag("env", "e", "environment", "")
	projectFlag := flags.NewStringFlag("project", "p", "project", "")
	fileFlag := flags.NewStringFlag("file", "f", "path to the YAML file containing the resource data to be updated", "")

	cmd := &cobra.Command{
		Use:   "update <resource> <id>",
		Short: "Update a specific resource using data from a YAML file",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			err := service.UpdateHandler(
				*envFlag.Value,
				*projectFlag.Value,
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
	core.RegisterCommand(updateCmd)
}
