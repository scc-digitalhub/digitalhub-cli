// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0
package cmd

import (
	"dhcli/handlers/environment"
	"dhcli/pkg"
	"dhcli/pkg/flags"
	"log"

	"github.com/spf13/cobra"
)

var registerCmd = func() *cobra.Command {
	envFlag := flags.NewStringFlag("env", "e", "environment", "")
	forceFlag := flags.NewBoolFlag("force", "f", "override existing environment with different endpoint", false)

	cmd := &cobra.Command{
		Use:   "register <endpoint>",
		Short: "Register the configuration of a core instance",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			endpoint := args[0]

			if err := environment.RegisterHandler(*envFlag.Value, endpoint, *forceFlag.Value); err != nil {
				log.Fatalf("Registration failed: %v", err)
			}
		},
	}

	flags.AddFlag(cmd, &envFlag)
	flags.AddFlag(cmd, &forceFlag)

	return cmd
}()

func init() {
	pkg.RegisterCommand(registerCmd)
}
