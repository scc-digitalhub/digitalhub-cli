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

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in to a given environment",
	Long:  "Authenticate the user using OAuth2 PKCE flow with the specified environment.",
	Run: func(cmd *cobra.Command, args []string) {

		if err := service.LoginHandler(); err != nil {
			log.Fatalf("Login failed: %v", err)
		}
	},
}

func init() {
	flags.AddCommonFlags(loginCmd, "env")
	core.RegisterCommand(loginCmd)
}
