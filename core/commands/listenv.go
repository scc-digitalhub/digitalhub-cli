// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"dhcli/core"
	"dhcli/core/service"

	"github.com/spf13/cobra"
)

var listEnvCmd = &cobra.Command{
	Use:   "list-env",
	Short: "List available environments",
	Run: func(cmd *cobra.Command, args []string) {
		service.ListEnvHandler()
	},
}

func init() {
	core.RegisterCommand(listEnvCmd)
}
