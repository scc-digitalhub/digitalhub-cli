// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"log"

	"dhcli/handlers/adapter"
	"dhcli/handlers/utils"
	"dhcli/pkg"
	"dhcli/pkg/flags"

	"github.com/spf13/cobra"
)

var eventsCmd = func() *cobra.Command {
	envFlag := flags.NewStringFlag("env", "e", "environment", "")
	outFlag := flags.NewStringFlag("out", "o", "output format (short, json, yaml)", "")
	projectFlag := flags.NewStringFlag("project", "p", "Project name (filters events client-side)", "")
	nameFlag := flags.NewStringFlag("name", "n", "Resource name (filters events client-side)", "")

	cmd := &cobra.Command{
		Use:   "events <resource> [id]",
		Short: "Stream real-time push events for a resource",
		Long: `Connect to the platform's STOMP/WebSocket broker and stream
push notifications for the given resource type.

With an optional ID argument the subscription is scoped to that specific object:

  dhcli events runs
  dhcli -p myproject events runs
  dhcli events runs <run-id>
  dhcli -p myproject events runs <run-id> -o json

Press Ctrl+C to disconnect.`,
		Args: cobra.RangeArgs(1, 2),
		Run: func(cmd *cobra.Command, args []string) {
			resource := args[0]
			id := ""
			if len(args) == 2 {
				id = args[1]
			}

			project := utils.ResolveProject(*projectFlag.Value)

			if err := adapter.EventsHandler(
				*envFlag.Value,
				*outFlag.Value,
				project,
				*nameFlag.Value,
				resource,
				id,
			); err != nil {
				log.Fatalf("Failed: %v", err)
			}
		},
	}

	flags.AddFlag(cmd, &envFlag)
	flags.AddFlag(cmd, &outFlag)
	flags.AddFlag(cmd, &projectFlag)
	flags.AddFlag(cmd, &nameFlag)

	return cmd
}()

func init() {
	pkg.RegisterCommand(eventsCmd)
}
