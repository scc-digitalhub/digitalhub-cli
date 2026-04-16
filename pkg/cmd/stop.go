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

var stopCmd = func() *cobra.Command {
	envFlag := flags.NewStringFlag("env", "e", "environment", "")
	projectFlag := flags.NewStringFlag("project", "p", "Mandatory", "")

	cmd := &cobra.Command{
		Use:   "stop <id>",
		Short: "Stop a run",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			project := utils.ResolveProject(*projectFlag.Value)
			err := adapter.StopHandler(
				*envFlag.Value,
				project,
				args[0],
			)
			if err != nil {
				log.Fatalf("Failed: %v", err)
			}
		},
	}

	flags.AddFlag(cmd, &envFlag)
	flags.AddFlag(cmd, &projectFlag)

	return cmd
}()

func init() {
	pkg.RegisterCommand(stopCmd)
}
