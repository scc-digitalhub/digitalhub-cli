// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"dhcli/core"
	"dhcli/core/facade"

	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove <environment>",
	Short: "Remove an environment from the configuration",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		facade.RemoveHandler(args[0])
	},
}

func init() {
	core.RegisterCommand(removeCmd)
}
