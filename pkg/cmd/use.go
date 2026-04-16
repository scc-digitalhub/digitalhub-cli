// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"dhcli/pkg"
	"dhcli/handlers/environment"

	"github.com/spf13/cobra"
)

var useCmd = &cobra.Command{
	Use:   "use <environment>",
	Short: "Sets the default environment",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		environment.UseHandler(args[0])
	},
}

func init() {
	pkg.RegisterCommand(useCmd)
}
