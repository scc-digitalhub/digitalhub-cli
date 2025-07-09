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

var uploadFlag = flags.SpecificCommandFlag{}

var uploadCmd = &cobra.Command{
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
			flags.CommonFlag.EnvFlag,
			uploadFlag.InputFlag,
			flags.CommonFlag.ProjectFlag,
			args[0],
			id,
			flags.CommonFlag.NameFlag,
		)
		if err != nil {
			log.Fatalf("Upload failed: %v", err)
		}
	},
}

func init() {

	flags.AddCommonFlags(uploadCmd, "env", "project", "name")

	uploadCmd.Flags().StringVarP(&uploadFlag.InputFlag, "file", "f", "", "input filename or directory")

	core.RegisterCommand(uploadCmd)
}
