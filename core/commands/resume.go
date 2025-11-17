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

var resumeCmd = func() *cobra.Command {
	envFlag := flags.NewStringFlag("env", "e", "environment", "")
	projectFlag := flags.NewStringFlag("project", "p", "Mandatory", "")

	cmd := &cobra.Command{
		Use:   "resume <resource> <id>",
		Short: "Resume a resource",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			err := adapter.ResumeHandler(
				*envFlag.Value,
				*projectFlag.Value,
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

	return cmd
}()

func init() {
	core.RegisterCommand(resumeCmd)
}
