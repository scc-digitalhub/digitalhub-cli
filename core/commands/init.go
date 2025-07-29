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

var initCmd = func() *cobra.Command {
	// Local flag declarations
	envFlag := flags.NewStringFlag("env", "e", "environment", "")
	preFlag := flags.NewBoolFlag("pre", "", "Include pre-release versions when installing", false)

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Install python packages for an environment",
		Run: func(cmd *cobra.Command, args []string) {
			if err := service.InitEnvironmentHandler(*preFlag.Value); err != nil {
				log.Fatalf("Init failed: %v", err)
			}
		},
	}

	flags.AddFlag(cmd, &envFlag)
	flags.AddFlag(cmd, &preFlag)

	return cmd
}()

func init() {
	core.RegisterCommand(initCmd)
}
