// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"dhcli/core"
	"dhcli/core/flags"
	"dhcli/core/service"
	"errors"
	"log"

	"github.com/spf13/cobra"
)

var getCmd = func() *cobra.Command {
	// Local flag declarations
	envFlag := flags.NewStringFlag("env", "e", "environment", "")
	outFlag := flags.NewStringFlag("out", "o", "output format (short, json, yaml)", "")
	projectFlag := flags.NewStringFlag("project", "p", "project", "")
	nameFlag := flags.NewStringFlag("name", "n", "name", "")

	cmd := &cobra.Command{
		Use:   "get <resource> [id]",
		Short: "Retrieve a resource",
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

			err := service.GetHandler(
				*envFlag.Value,
				*outFlag.Value,
				*projectFlag.Value,
				*nameFlag.Value,
				args[0],
				id,
			)
			if err != nil {
				log.Fatalf("Get failed: %v", err)
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
	core.RegisterCommand(getCmd)
}
