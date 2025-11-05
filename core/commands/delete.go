// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"dhcli/core"
	"dhcli/core/flags"
	"dhcli/core/service/adapter"
	"errors"
	"log"

	"github.com/spf13/cobra"
)

var deleteCmd = func() *cobra.Command {
	// Declare flags locally with proper initialization
	envFlag := flags.NewStringFlag("env", "e", "environment", "")
	projectFlag := flags.NewStringFlag("project", "p", "Mandatory for resources other than projects", "")
	nameFlag := flags.NewStringFlag("name", "n", "Alternative to id, will delete all versions of resource", "")
	confirmFlag := flags.NewBoolFlag("confirm", "y", "Skips the deletion confirmation prompt", false)
	cascadeFlag := flags.NewBoolFlag("cascade", "c", "If set, also deletes related resources (for projects)", false)

	cmd := &cobra.Command{
		Use:   "delete <resource> [<id>]",
		Short: "Delete a resource by ID or name",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 || len(args) > 2 {
				return errors.New("requires 1 or 2 arguments: <resource> [<id>]")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			id := ""
			if len(args) > 1 {
				id = args[1]
			}

			err := adapter.DeleteHandler(
				*envFlag.Value,
				*projectFlag.Value,
				*nameFlag.Value,
				*confirmFlag.Value,
				*cascadeFlag.Value,
				args[0],
				id,
			)

			if err != nil {
				log.Fatalf("Delete failed: %v", err)
			}
		},
	}

	// Add flags to cmd
	flags.AddFlag(cmd, &envFlag)
	flags.AddFlag(cmd, &projectFlag)
	flags.AddFlag(cmd, &nameFlag)
	flags.AddFlag(cmd, &confirmFlag)
	flags.AddFlag(cmd, &cascadeFlag)

	return cmd
}()

func init() {
	core.RegisterCommand(deleteCmd)
}
