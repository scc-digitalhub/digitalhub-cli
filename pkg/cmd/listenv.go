// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"dhcli/pkg"
	"dhcli/handlers/environment"

	"github.com/spf13/cobra"
)

var listEnvCmd = &cobra.Command{
	Use:   "list-env",
	Short: "List available environments",
	Run: func(cmd *cobra.Command, args []string) {
		environment.ListEnvHandler()
	},
}

func init() {
	core.RegisterCommand(listEnvCmd)
}
