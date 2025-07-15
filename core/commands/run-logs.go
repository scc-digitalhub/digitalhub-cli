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

var runLogsCmd = func() *cobra.Command {
	envFlag := flags.NewStringFlag("env", "e", "environment", "")

	cmd := &cobra.Command{
		Use:   "run-logs <project> <id>",
		Short: "Read run logs",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			err := service.RunLogsHandler(
				*envFlag.Value,
				args[0],
				args[1],
			)
			if err != nil {
				log.Fatalf("Failed: %v", err)
			}
		},
	}

	flags.AddFlag(cmd, &envFlag)

	return cmd
}()

func init() {
	core.RegisterCommand(runLogsCmd)
}
