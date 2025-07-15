// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"dhcli/core"
	"dhcli/core/flags"
	"dhcli/core/service"
	"errors"
	"github.com/spf13/cobra"
	"log"
)

var uploadCmd = func() *cobra.Command {
	envFlag := flags.NewStringFlag("env", "e", "environment", "")
	projectFlag := flags.NewStringFlag("project", "p", "project", "")
	nameFlag := flags.NewStringFlag("name", "n", "name", "")
	inputFlag := flags.NewStringFlag("file", "f", "input filename or directory", "")

	cmd := &cobra.Command{
		Use:   "upload <resource>",
		Short: "Upload a resource on the S3 aws",
		Long:  "Upload an artifact from ........................",
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

			err := service.UploadHandler(
				*envFlag.Value,
				*inputFlag.Value,
				*projectFlag.Value,
				args[0],
				id,
				*nameFlag.Value,
			)
			if err != nil {
				log.Fatalf("Upload failed: %v", err)
			}
		},
	}

	flags.AddFlag(cmd, &envFlag)
	flags.AddFlag(cmd, &projectFlag)
	flags.AddFlag(cmd, &nameFlag)
	flags.AddFlag(cmd, &inputFlag)

	return cmd
}()

func init() {
	core.RegisterCommand(uploadCmd)
}
