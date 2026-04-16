// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"dhcli/handlers/adapter"
	"dhcli/pkg"
	"dhcli/pkg/flags"
	"log"

	"github.com/spf13/cobra"
)

var servicesCmd = func() *cobra.Command {
	// Declare local flags using generic constructors
	envFlag := flags.NewStringFlag("env", "e", "environment", "")
	outFlag := flags.NewStringFlag("out", "o", "output format (short, json, yaml)", "short")
	projectFlag := flags.NewStringFlag("project", "p", "Mandatory for listing services", "")
	nameFlag := flags.NewStringFlag("name", "n", "If specified, all versions of the service will be listed", "")

	kindFlag := flags.NewStringFlag("kind", "k", "Filter by kind", "")
	stateFlag := flags.NewStringFlag("state", "s", "Filter by state", "")

	cmd := &cobra.Command{
		Use:   "services",
		Short: "List services (runs with action=serve)",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if err := adapter.ListServicesHandler(
				*envFlag.Value,
				*outFlag.Value,
				*projectFlag.Value,
				*nameFlag.Value,
				*kindFlag.Value,
				*stateFlag.Value,
			); err != nil {
				log.Fatalf("Services list failed: %v", err)
			}
		},
	}

	// Add common flags
	flags.AddFlag(cmd, &envFlag)
	flags.AddFlag(cmd, &outFlag)
	flags.AddFlag(cmd, &projectFlag)
	flags.AddFlag(cmd, &nameFlag)

	// Add specific flags
	flags.AddFlag(cmd, &kindFlag)
	flags.AddFlag(cmd, &stateFlag)

	return cmd
}()

func init() {
	pkg.RegisterCommand(servicesCmd)
}
