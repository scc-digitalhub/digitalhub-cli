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

var downloadCmd = func() *cobra.Command {
	// Declare flags locally
	envFlag := flags.NewStringFlag("env", "e", "environment", "")
	projectFlag := flags.NewStringFlag("project", "p", "Mandatory for resources other than projects", "")
	nameFlag := flags.NewStringFlag("name", "n", "Alternative to id, will download latest version", "")
	destinationFlag := flags.NewStringFlag("destination", "d", "output filename or directory", "")
	outFlag := flags.NewStringFlag("out", "o", "Output format (short, json, yaml)", "")
	verboseFlag := flags.NewBoolFlag("verbose", "v", "Verbose progress/logging", false)

	cmd := &cobra.Command{
		Use:   "download <resource> [<id>]",
		Short: "Download a resource from the S3 aws",
		Long:  "Download a resource from S3 aws",
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

			if err := service.DownloadHandler(
				*envFlag.Value,
				*destinationFlag.Value,
				*outFlag.Value,
				*projectFlag.Value,
				*nameFlag.Value,
				args[0],
				id,
				*verboseFlag.Value,
			); err != nil {
				log.Fatalf("Download failed: %v", err)
			}
		},
	}

	// Add flags
	flags.AddFlag(cmd, &envFlag)
	flags.AddFlag(cmd, &projectFlag)
	flags.AddFlag(cmd, &nameFlag)
	flags.AddFlag(cmd, &outFlag)
	flags.AddFlag(cmd, &destinationFlag)
	flags.AddFlag(cmd, &verboseFlag)

	return cmd
}()

func init() {
	core.RegisterCommand(downloadCmd)
}
