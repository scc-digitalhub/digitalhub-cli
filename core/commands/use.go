// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"dhcli/core"
	"dhcli/core/facade"

	"github.com/spf13/cobra"
)

var useCmd = &cobra.Command{
	Use:   "use <environment>",
	Short: "Sets the default environment",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		facade.UseHandler(args[0])
	},
}

func init() {
	core.RegisterCommand(useCmd)
}
