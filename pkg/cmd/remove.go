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

var removeCmd = &cobra.Command{
	Use:   "remove <environment>",
	Short: "Remove an environment from the configuration",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := environment.RemoveHandler(args[0]); err != nil {
			log.Fatalf("Remove failed: %v", err)
		}
	},
}

func init() {
	pkg.RegisterCommand(removeCmd)
}
