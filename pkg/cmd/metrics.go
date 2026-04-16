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

var metricsCmd = func() *cobra.Command {
	envFlag := flags.NewStringFlag("env", "e", "environment", "")
	projectFlag := flags.NewStringFlag("project", "p", "Mandatory", "")
	containerFlag := flags.NewStringFlag("container", "c", "Container ID", "")

	cmd := &cobra.Command{
		Use:   "metrics <resource> <id>",
		Short: "Read metrics",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			project := utils.ResolveProject(*projectFlag.Value)
			err := adapter.MetricsHandler(
				*envFlag.Value,
				project,
				*containerFlag.Value,
				args[0],
				args[1],
			)
			if err != nil {
				log.Fatalf("Failed: %v", err)
			}
		},
	}

	flags.AddFlag(cmd, &envFlag)
	flags.AddFlag(cmd, &projectFlag)
	flags.AddFlag(cmd, &containerFlag)

	return cmd
}()

func init() {
	pkg.RegisterCommand(metricsCmd)
}
