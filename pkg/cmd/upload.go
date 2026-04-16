// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"dhcli/handlers/adapter"
	"dhcli/pkg"
	"dhcli/pkg/flags"
	"errors"
	"log"

	"dhcli/handlers/utils"

	"github.com/spf13/cobra"
)

var uploadCmd = func() *cobra.Command {
	envFlag := flags.NewStringFlag("env", "e", "environment", "")
	projectFlag := flags.NewStringFlag("project", "p", "Mandatory for resources other than projects", "")
	nameFlag := flags.NewStringFlag("name", "n", "Mandatory when creating a new artifact", "")
	inputFlag := flags.NewStringFlag("file", "f", "Input filename or directory; mandatory", "")
	verboseFlag := flags.NewBoolFlag("verbose", "v", "Verbose progress/logging", false)

	cmd := &cobra.Command{
		Use:   "upload <resource> [<id>]",
		Short: "Upload a resource to S3",
		Long:  "Upload a file or directory to S3, optionally creating a new artifact when ID is omitted.",
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

			project := utils.ResolveProject(*projectFlag.Value)
			err := adapter.UploadHandler(
				*envFlag.Value,
				*inputFlag.Value,
				project,
				args[0],
				id,
				*nameFlag.Value,
				*verboseFlag.Value,
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
	flags.AddFlag(cmd, &verboseFlag)

	return cmd
}()

func init() {
	pkg.RegisterCommand(uploadCmd)
}
