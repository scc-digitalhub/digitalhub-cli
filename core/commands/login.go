// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"dhcli/core"
	"dhcli/core/service"
	"log"

	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login [environment]",
	Short: "Log in to a given environment",
	Long:  "Authenticate the user using OAuth2 PKCE flow with the specified environment.",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var environment string
		if len(args) > 0 {
			environment = args[0]
		}

		if err := service.LoginHandler(environment); err != nil {
			log.Fatalf("Login failed: %v", err)
		}
	},
}

func init() {
	core.RegisterCommand(loginCmd)
}
