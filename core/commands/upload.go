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
	Short: "upload a resource on the S3 aws",
	Long:  "Upload an artifact from ........................",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("requires exactly 1 argument: <resource>")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {

		err := service.UploadHandler(
			flags.CommonFlag.EnvFlag,
			uploadFlag.InputFlag,
			flags.CommonFlag.ProjectFlag,
			uploadFlag.IdFlag,
			args[0],
			flags.CommonFlag.NameFlag,
		)
		if err != nil {
			log.Fatalf("Upload failed: %v", err)
		}
	},
}

func init() {
	flags.AddCommonFlags(uploadCmd, "env", "project", "name")

	uploadCmd.Flags().StringVarP(&uploadFlag.InputFlag, "input", "i", "", "input filename or directory")
	uploadCmd.Flags().StringVarP(&uploadFlag.IdFlag, "key", "k", "", "artifact id to use for the upload")

	core.RegisterCommand(uploadCmd)
}
