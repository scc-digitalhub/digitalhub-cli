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

var configCmd = func() *cobra.Command {
	envFlag := flags.NewStringFlag("env", "e", "environment", "")
	outFlag := flags.NewStringFlag("out", "o", "output format (short, json, yaml)", "")
	providerFlag := flags.NewStringFlag("provider", "p", "filter config by provider (e.g. dhcore, trino, s3)", "")

	cmd := &cobra.Command{
		Use:   "config",
		Short: "Print current environment configuration (non-secret values)",
		Run: func(cmd *cobra.Command, args []string) {
			_ = envFlag // env is handled by PersistentPreRunE
			err := config.ConfigHandler(*outFlag.Value, *providerFlag.Value)
			if err != nil {
				log.Fatalf("Config failed: %v", err)
			}
		},
	}

	flags.AddFlag(cmd, &envFlag)
	flags.AddFlag(cmd, &outFlag)
	flags.AddFlag(cmd, &providerFlag)

	return cmd
}()

func init() {
	pkg.RegisterCommand(configCmd)
}
