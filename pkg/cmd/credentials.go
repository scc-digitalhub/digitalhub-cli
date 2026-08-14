// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"log"

	"dhcli/handlers/config"
	"dhcli/pkg"
	"dhcli/pkg/flags"

	"github.com/spf13/cobra"
)

var credentialsCmd = func() *cobra.Command {
	envFlag := flags.NewStringFlag("env", "e", "environment", "")
	outFlag := flags.NewStringFlag("out", "o", "output format (short, json, yaml)", "")
	providerFlag := flags.NewStringFlag("provider", "p", "filter credentials by provider (e.g. dhcore, trino, s3)", "")

	cmd := &cobra.Command{
		Use:   "credentials",
		Short: "Print current environment credentials (secret values)",
		Run: func(cmd *cobra.Command, args []string) {
			_ = envFlag // env is handled by PersistentPreRunE
			err := config.CredentialsHandler(*outFlag.Value, *providerFlag.Value)
			if err != nil {
				log.Fatalf("Credentials failed: %v", err)
			}
		},
	}

	flags.AddFlag(cmd, &envFlag)
	flags.AddFlag(cmd, &outFlag)
	flags.AddFlag(cmd, &providerFlag)

	return cmd
}()

func init() {
	pkg.RegisterCommand(credentialsCmd)
}
