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

var refreshCmd = func() *cobra.Command {
	// Declare local env flag
	envFlag := flags.NewStringFlag("env", "e", "environment", "")

	cmd := &cobra.Command{
		Use:   "refresh",
		Short: "Refresh access token",
		Long:  "Refresh the access token of a given environment.",
		Run: func(cmd *cobra.Command, args []string) {
			if err := facade.RefreshHandler(); err != nil {
				log.Fatalf("Refresh failed: %v", err)
			}
		},
	}

	// Add local env flag
	flags.AddFlag(cmd, &envFlag)

	return cmd
}()

func init() {
	core.RegisterCommand(refreshCmd)
}
