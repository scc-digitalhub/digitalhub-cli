// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"log"

	"dhcli/handlers/environment"
	"dhcli/pkg"

	"github.com/spf13/cobra"
)

var useCmd = &cobra.Command{
	Use:   "use <environment>",
	Short: "Sets the default environment",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := environment.UseHandler(args[0]); err != nil {
			log.Fatalf("Use failed: %v", err)
		}
	},
}

func init() {
	pkg.RegisterCommand(useCmd)
}
