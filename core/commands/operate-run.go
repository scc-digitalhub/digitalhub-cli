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

var operateRunCmd = func() *cobra.Command {
	// Define local env flag
	envFlag := flags.NewStringFlag("env", "e", "environment", "")

	cmd := &cobra.Command{
		Use:   "operate-run <project> <id> <operation>",
		Short: "Perform an operation on a run",
		Args:  cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			err := service.OperateRunHandler(
				*envFlag.Value,
				args[0],
				args[1],
				args[2],
			)
			if err != nil {
				log.Fatalf("Failed: %v", err)
			}
		},
	}

	// Add the local env flag
	flags.AddFlag(cmd, &envFlag)

	return cmd
}()

func init() {
	core.RegisterCommand(operateRunCmd)
}
