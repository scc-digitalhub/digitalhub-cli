// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"dhcli/core"
	"dhcli/core/flags"
	"dhcli/core/service"
	"log"

	"github.com/spf13/cobra"
)

var listCmd = func() *cobra.Command {
	// Declare local flags using generic constructors
	envFlag := flags.NewStringFlag("env", "e", "environment", "")
	outFlag := flags.NewStringFlag("out", "o", "output format (short, json, yaml)", "short")
	projectFlag := flags.NewStringFlag("project", "p", "project", "")
	nameFlag := flags.NewStringFlag("name", "n", "name", "")

	kindFlag := flags.NewStringFlag("kind", "k", "kind", "")
	stateFlag := flags.NewStringFlag("state", "s", "state", "")

	cmd := &cobra.Command{
		Use:   "list <resource>",
		Short: "List resources",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := service.ListResourcesHandler(
				*envFlag.Value,
				*outFlag.Value,
				*projectFlag.Value,
				*nameFlag.Value,
				*kindFlag.Value,
				*stateFlag.Value,
				args[0],
			); err != nil {
				log.Fatalf("List failed: %v", err)
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
	core.RegisterCommand(listCmd)
}
