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

var (
	fileOrDirectoryFlag string
)

var downloadCmd = &cobra.Command{
	Use:   "download <resource> <id>",
	Short: "Download a resource from the S3 aws",
	Long:  "Download an artifact from ........................",
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
			flags.EnvFlag,
			fileOrDirectoryFlag,
			flags.ProjectFlag,
			flags.NameFlag,
			args[0],
			id,
			args[1:]); err != nil {
			log.Fatalf("Download failed: %v", err)
		}
	},
}

func init() {
	flags.AddCommonFlags(downloadCmd, "env", "project", "name")

	downloadCmd.Flags().StringVarP(&fileOrDirectoryFlag, "out", "o", "", "output filename or directory")

	core.RegisterCommand(downloadCmd)
}
