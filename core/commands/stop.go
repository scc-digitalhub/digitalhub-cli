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

var stopCmd = &cobra.Command{
	Use:   "stop <resource> <id>",
	Short: "Stop a resource",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		err := service.StopHandler(
			flags.CommonFlag.EnvFlag,
			flags.CommonFlag.ProjectFlag,
			args[0],
			args[1])

		if err != nil {
			log.Fatalf("Failed: %v", err)
		}
	},
}

func init() {
	flags.AddCommonFlags(stopCmd, "env", "project")

	core.RegisterCommand(stopCmd)
}
