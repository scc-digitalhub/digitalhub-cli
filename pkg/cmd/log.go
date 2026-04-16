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

var logCmd = func() *cobra.Command {
	envFlag := flags.NewStringFlag("env", "e", "environment", "")
	projectFlag := flags.NewStringFlag("project", "p", "Mandatory", "")
	containerFlag := flags.NewStringFlag("container", "c", "Container ID", "")
	followFlag := flags.NewBoolFlag("follow", "f", "Attach console and continue to refresh logs", false)

	cmd := &cobra.Command{
		Use:   "log <id>",
		Short: "Read logs",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			project := utils.ResolveProject(*projectFlag.Value)
			err := adapter.LogHandler(
				*envFlag.Value,
				project,
				*containerFlag.Value,
				*followFlag.Value,
				args[0],
			)

			if err != nil {
				log.Fatalf("Failed: %v", err)
			}
		},
	}

	flags.AddFlag(cmd, &envFlag)
	flags.AddFlag(cmd, &projectFlag)
	flags.AddFlag(cmd, &containerFlag)
	flags.AddFlag(cmd, &followFlag)

	return cmd
}()

func init() {
	pkg.RegisterCommand(logCmd)
}
