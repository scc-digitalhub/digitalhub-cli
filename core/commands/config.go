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

var configCmd = func() *cobra.Command {
	envFlag := flags.NewStringFlag("env", "e", "environment", "")
	outFlag := flags.NewStringFlag("out", "o", "output format (short, json, yaml)", "")

	cmd := &cobra.Command{
		Use:   "config",
		Short: "Print current environment configuration (non-secret values)",
		Run: func(cmd *cobra.Command, args []string) {
			_ = envFlag // env is handled by PersistentPreRunE
			err := facade.ConfigHandler(*outFlag.Value)
			if err != nil {
				log.Fatalf("Config failed: %v", err)
			}
		},
	}

	flags.AddFlag(cmd, &envFlag)
	flags.AddFlag(cmd, &outFlag)

	return cmd
}()

func init() {
	core.RegisterCommand(configCmd)
}
