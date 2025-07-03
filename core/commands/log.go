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

var logFlag = flags.SpecificCommandFlag{}

var logCmd = &cobra.Command{
	Use:   "log <resource> <id>",
	Short: "Read logs",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		err := service.LogHandler(
			flags.CommonFlag.EnvFlag,
			flags.CommonFlag.ProjectFlag,
			logFlag.ContainerFlag,
			logFlag.FollowFlag,
			args[0],
			args[1])

		if err != nil {
			log.Fatalf("Failed: %v", err)
		}
	},
}

func init() {
	flags.AddCommonFlags(logCmd, "env", "project")

	// Additional flags
	logCmd.Flags().StringVarP(&logFlag.ContainerFlag, "container", "c", "", "Container ID")
	logCmd.Flags().BoolVarP(&logFlag.FollowFlag, "follow", "f", false, "Attach console and continue to refresh logs")

	core.RegisterCommand(logCmd)
}
