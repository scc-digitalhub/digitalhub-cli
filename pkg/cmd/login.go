// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"dhcli/handlers/auth"
	"dhcli/pkg"
	"dhcli/pkg/flags"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var loginCmd = func() *cobra.Command {
	// Declare local flags
	envFlag := flags.NewStringFlag("env", "e", "environment", "")
	patFlag := flags.NewStringFlag("pat", "", "personal access token (non-interactive flow)", "")

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Log in to a given environment",
		Long:  "Authenticate the user using OAuth2 PKCE flow with the specified environment. Use --pat (or env DHCORE_PERSONAL_ACCESS_TOKEN / DHCORE_PAT) for non-interactive token exchange.",
		Run: func(cmd *cobra.Command, args []string) {
			pat := *patFlag.Value
			if pat == "" {
				pat = os.Getenv("DHCORE_PERSONAL_ACCESS_TOKEN")
			}
			if pat == "" {
				pat = os.Getenv("DHCORE_PAT")
			}

			if pat != "" {
				if err := auth.PatLoginHandler(pat); err != nil {
					log.Fatalf("Login failed: %v", err)
				}
				return
			}

			if err := auth.LoginHandler(); err != nil {
				log.Fatalf("Login failed: %v", err)
			}
		},
	}

	// Add local flags
	flags.AddFlag(cmd, &envFlag)
	flags.AddFlag(cmd, &patFlag)

	return cmd
}()

func init() {
	pkg.RegisterCommand(loginCmd)
}
