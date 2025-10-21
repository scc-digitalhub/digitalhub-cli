// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"dhcli/core"
	"dhcli/core/flags"
	"dhcli/core/service/adapter"
	"log"

	"github.com/spf13/cobra"
)

var loginCmd = func() *cobra.Command {
	// Declare local env flag
	envFlag := flags.NewStringFlag("env", "e", "environment", "")

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Log in to a given environment",
		Long:  "Authenticate the user using OAuth2 PKCE flow with the specified environment.",
		Run: func(cmd *cobra.Command, args []string) {
			// Pass the dereferenced envFlag value if needed in service.LoginHandler (adjust if required)
			if err := adapter.LoginHandler(); err != nil {
				log.Fatalf("Login failed: %v", err)
			}
		},
	}

	// Add local env flag
	flags.AddFlag(cmd, &envFlag)

	return cmd
}()

func init() {
	core.RegisterCommand(loginCmd)
}
