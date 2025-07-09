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

var runFlag = flags.SpecificCommandFlag{}

var runCmd = &cobra.Command{
	Use:   "run <task>",
	Short: "Runs a function",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := service.RunHandler(
			flags.CommonFlag.EnvFlag,
			flags.CommonFlag.ProjectFlag,
			runFlag.FunctionNameFlag,
			runFlag.FunctionIdFlag,
			runFlag.FilePathFlag,
			args[0])
		if err != nil {
			log.Fatalf("Run failed: %v", err)
		}
	},
}

func init() {
	flags.AddCommonFlags(runCmd, "env", "project")

	// Additional flags
	runCmd.Flags().StringVarP(&runFlag.FunctionNameFlag, "fn-name", "n", "", "name of the function to run")
	runCmd.Flags().StringVarP(&runFlag.FunctionIdFlag, "fn-id", "i", "", "ID of the function to run")
	runCmd.Flags().StringVarP(&runFlag.FilePathFlag, "file", "f", "", "path to a YAML file containing the resource definition")

	core.RegisterCommand(runCmd)
}
