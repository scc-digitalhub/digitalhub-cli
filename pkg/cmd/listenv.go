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

var listEnvCmd = &cobra.Command{
	Use:   "list-env",
	Short: "List available environments",
	Run: func(cmd *cobra.Command, args []string) {
		if err := environment.ListEnvHandler(); err != nil {
			log.Fatalf("List environments failed: %v", err)
		}
	},
}

func init() {
	pkg.RegisterCommand(listEnvCmd)
}
