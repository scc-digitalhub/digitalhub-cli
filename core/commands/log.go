// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"dhcli/core"
	"dhcli/core/flags"
	"dhcli/core/service/adapter"
	"log"

	"github.com/spf13/cobra"
)

var logCmd = func() *cobra.Command {
	envFlag := flags.NewStringFlag("env", "e", "environment", "")
	projectFlag := flags.NewStringFlag("project", "p", "Mandatory", "")
	containerFlag := flags.NewStringFlag("container", "c", "Container ID", "")
	followFlag := flags.NewBoolFlag("follow", "f", "Attach console and continue to refresh logs", false)

	cmd := &cobra.Command{
		Use:   "log <resource> <id>",
		Short: "Read logs",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			err := adapter.LogHandler(
				*envFlag.Value,
				*projectFlag.Value,
				*containerFlag.Value,
				*followFlag.Value,
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
	flags.AddFlag(cmd, &followFlag)

	return cmd
}()

func init() {
	core.RegisterCommand(logCmd)
}
