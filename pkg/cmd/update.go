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

var updateCmd = func() *cobra.Command {
	envFlag := flags.NewStringFlag("env", "e", "environment", "")
	outputFlag := flags.NewStringFlag("output", "o", "output format (short, json, yaml)", "")
	projectFlag := flags.NewStringFlag("project", "p", "Mandatory for resources other than projects", "")
	fileFlag := flags.NewStringFlag("file", "f", "path to the YAML file containing the resource data to be updated", "")
	sourceFileFlag := flags.NewStringFlag("source-file", "", "Path to a local source file (.py, .sql) to Base64-encode into spec.source.base64", "")
	valueFlag := flags.NewStringFlag("value", "", "Secret value (visible in process arguments; prefer --value-stdin)", "")
	valueStdinFlag := flags.NewBoolFlag("value-stdin", "", "Read the secret value from standard input", false)

	cmd := &cobra.Command{
		Use:   "update <resource> <id>",
		Short: "Update a specific resource using data from a YAML file (or --value for secrets)",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			secretValue, secretSet, err := readSecretValue(cmd, args[0], *valueFlag.Value, *valueStdinFlag.Value)
			if err != nil {
				log.Fatalf("Update failed: %v", err)
			}
			project := utils.ResolveProject(*projectFlag.Value)
			err = adapter.UpdateWithOptionsHandler(adapter.UpdateOptions{
				Env:            *envFlag.Value,
				Output:         *outputFlag.Value,
				Project:        project,
				FilePath:       *fileFlag.Value,
				SourceFile:     *sourceFileFlag.Value,
				SecretValue:    secretValue,
				SecretValueSet: secretSet,
				Resource:       args[0],
				ID:             args[1],
			})
			if err != nil {
				log.Fatalf("Update failed: %v", err)
			}
		},
	}

	flags.AddFlag(cmd, &envFlag)
	flags.AddFlag(cmd, &outputFlag)
	flags.AddFlag(cmd, &projectFlag)
	flags.AddFlag(cmd, &fileFlag)
	flags.AddFlag(cmd, &sourceFileFlag)
	flags.AddFlag(cmd, &valueFlag)
	flags.AddFlag(cmd, &valueStdinFlag)

	return cmd
}()

func init() {
	pkg.RegisterCommand(updateCmd)
}
