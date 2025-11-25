// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0
package commands

import (
	"dhcli/core"
	"dhcli/core/facade"
	"dhcli/core/flags"
	"log"

	"github.com/spf13/cobra"
)

var registerCmd = func() *cobra.Command {
	envFlag := flags.NewStringFlag("env", "e", "environment", "")

	cmd := &cobra.Command{
		Use:   "register <endpoint>",
		Short: "Register the configuration of a core instance",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			endpoint := args[0]

			if err := facade.RegisterHandler(*envFlag.Value, endpoint); err != nil {
				log.Fatalf("Registration failed: %v", err)
			}
		},
	}

	flags.AddFlag(cmd, &envFlag)

	return cmd
}()

func init() {
	core.RegisterCommand(registerCmd)
}
