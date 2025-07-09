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

var initFlag = flags.SpecificCommandFlag{}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Install python packages for an environment",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := service.InitEnvironmentHandler(initFlag.PreFlag); err != nil {
			log.Fatalf("Init failed: %v", err)
		}
	},
}

func init() {
	flags.AddCommonFlags(initCmd, "env")
	initCmd.Flags().BoolVar(&initFlag.PreFlag, "pre", false, "Include pre-release versions when installing")
	core.RegisterCommand(initCmd)
}
