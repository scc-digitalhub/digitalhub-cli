// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"dhcli/pkg"
	"dhcli/pkg/flags"
	"dhcli/handlers/resources"
	"log"

	"github.com/spf13/cobra"
)

var credentialsCmd = func() *cobra.Command {
	envFlag := flags.NewStringFlag("env", "e", "environment", "")
	outFlag := flags.NewStringFlag("out", "o", "output format (short, json, yaml)", "")

	cmd := &cobra.Command{
		Use:   "credentials",
		Short: "Print current environment credentials (secret values)",
		Run: func(cmd *cobra.Command, args []string) {
			_ = envFlag // env is handled by PersistentPreRunE
			err := resources.CredentialsHandler(*outFlag.Value)
			if err != nil {
				log.Fatalf("Credentials failed: %v", err)
			}
		},
	}

	flags.AddFlag(cmd, &envFlag)
	flags.AddFlag(cmd, &outFlag)

	return cmd
}()

func init() {
	core.RegisterCommand(credentialsCmd)
}
