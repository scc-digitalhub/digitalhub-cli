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

var runCmd = func() *cobra.Command {
	envFlag := flags.NewStringFlag("env", "e", "environment", "")
	outputFlag := flags.NewStringFlag("output", "o", "output format (short, json, yaml)", "")
	projectFlag := flags.NewStringFlag("project", "p", "Mandatory", "")
	fnNameFlag := flags.NewStringFlag("fn-name", "n", "name of the function or workflow to run, alternative to Id", "")
	fnIDFlag := flags.NewStringFlag("fn-id", "i", "Id of the function or workflow to run, alternative to name", "")
	wfNameFlag := flags.NewStringFlag("wf-name", "", "name of the workflow to run (alias for -n)", "")
	wfIDFlag := flags.NewStringFlag("wf-id", "", "Id of the workflow to run (alias for -i)", "")
	filePathFlag := flags.NewStringFlag("file", "f", "path to a YAML file containing the run specification", "")

	cmd := &cobra.Command{
		Use:   "run <task>",
		Short: "Runs a function or workflow task (<kind>+<action>)",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			project := utils.ResolveProject(*projectFlag.Value)
			name := *fnNameFlag.Value
			id := *fnIDFlag.Value
			isWorkflow := false
			if *wfNameFlag.Value != "" {
				name = *wfNameFlag.Value
				isWorkflow = true
			}
			if *wfIDFlag.Value != "" {
				id = *wfIDFlag.Value
				isWorkflow = true
			}

			err := adapter.RunWithOptionsHandler(adapter.RunOptions{
				Env:        *envFlag.Value,
				Output:     *outputFlag.Value,
				Project:    project,
				Name:       name,
				ID:         id,
				FilePath:   *filePathFlag.Value,
				Task:       args[0],
				IsWorkflow: isWorkflow,
			})
			if err != nil {
				log.Fatalf("Run failed: %v", err)
			}
		},
	}

	flags.AddFlag(cmd, &envFlag)
	flags.AddFlag(cmd, &outputFlag)
	flags.AddFlag(cmd, &projectFlag)
	flags.AddFlag(cmd, &fnNameFlag)
	flags.AddFlag(cmd, &fnIDFlag)
	flags.AddFlag(cmd, &wfNameFlag)
	flags.AddFlag(cmd, &wfIDFlag)
	flags.AddFlag(cmd, &filePathFlag)

	return cmd
}()

func init() {
	pkg.RegisterCommand(runCmd)
}
