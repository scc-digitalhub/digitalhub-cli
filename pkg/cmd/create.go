// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"io"
	"log"

	"dhcli/handlers/adapter"
	"dhcli/handlers/utils"
	"dhcli/pkg"
	"dhcli/pkg/flags"

	"github.com/spf13/cobra"
)

var createCmd = func() *cobra.Command {
	envFlag := flags.NewStringFlag("env", "e", "environment", "")
	outputFlag := flags.NewStringFlag("output", "o", "output format (short, json, yaml)", "")
	projectFlag := flags.NewStringFlag("project", "p", "Mandatory for resources other than projects", "")
	nameFlag := flags.NewStringFlag("name", "n", "Resource name (projects, secrets, and triggers may be created without a file)", "")
	resetIdFlag := flags.NewBoolFlag("reset-id", "r", "If set, removes the id field from the file to ensure the server assigns a new one", false)
	fileFlag := flags.NewStringFlag("file", "f", "Path to a YAML file containing the resource definition", "")
	sourceFileFlag := flags.NewStringFlag("source-file", "", "Path to a local source file (.py, .sql) to Base64-encode into spec.source.base64", "")
	valueFlag := flags.NewStringFlag("value", "", "Secret value (visible in process arguments; prefer --value-stdin)", "")
	valueStdinFlag := flags.NewBoolFlag("value-stdin", "", "Read the secret value from standard input", false)
	taskFlag := flags.NewStringFlag("task", "", "Task kind (<kind+action>) when creating a trigger", "")
	targetNameFlag := flags.NewStringFlag("target-name", "", "Target function or workflow name when creating a trigger", "")
	fnNameFlag := flags.NewStringFlag("fn-name", "", "Target function name when creating a trigger", "")
	wfNameFlag := flags.NewStringFlag("wf-name", "", "Target workflow name when creating a trigger", "")
	scheduleFlag := flags.NewStringFlag("schedule", "", "Cron schedule expression when creating a scheduler trigger", "")

	cmd := &cobra.Command{
		Use:   "create <resource>",
		Short: "Creates a new resource from a YAML file (or flags for projects, secrets, and triggers)",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			secretValue, secretSet, err := readSecretValue(cmd, args[0], *valueFlag.Value, *valueStdinFlag.Value)
			if err != nil {
				log.Fatalf("Create failed: %v", err)
			}
			project := utils.ResolveProject(*projectFlag.Value)
			targetName := *targetNameFlag.Value
			if targetName == "" {
				if *fnNameFlag.Value != "" {
					targetName = *fnNameFlag.Value
				} else if *wfNameFlag.Value != "" {
					targetName = *wfNameFlag.Value
				}
			}
			err = adapter.CreateWithOptionsHandler(adapter.CreateOptions{
				Env:            *envFlag.Value,
				Output:         *outputFlag.Value,
				Project:        project,
				Name:           *nameFlag.Value,
				FilePath:       *fileFlag.Value,
				SourceFile:     *sourceFileFlag.Value,
				SecretValue:    secretValue,
				SecretValueSet: secretSet,
				Task:           *taskFlag.Value,
				TargetName:     targetName,
				Schedule:       *scheduleFlag.Value,
				ResetID:        *resetIdFlag.Value,
				Resource:       args[0],
			})
			if err != nil {
				log.Fatalf("Create failed: %v", err)
			}
		},
	}

	flags.AddFlag(cmd, &envFlag)
	flags.AddFlag(cmd, &outputFlag)
	flags.AddFlag(cmd, &projectFlag)
	flags.AddFlag(cmd, &nameFlag)
	flags.AddFlag(cmd, &resetIdFlag)
	flags.AddFlag(cmd, &fileFlag)
	flags.AddFlag(cmd, &sourceFileFlag)
	flags.AddFlag(cmd, &valueFlag)
	flags.AddFlag(cmd, &valueStdinFlag)
	flags.AddFlag(cmd, &taskFlag)
	flags.AddFlag(cmd, &targetNameFlag)
	flags.AddFlag(cmd, &fnNameFlag)
	flags.AddFlag(cmd, &wfNameFlag)
	flags.AddFlag(cmd, &scheduleFlag)

	return cmd
}()

func readSecretValue(cmd *cobra.Command, resource, value string, fromStdin bool) (string, bool, error) {
	valueSet := cmd.Flags().Changed("value")
	if fromStdin && valueSet {
		return "", false, fmt.Errorf("--value and --value-stdin cannot be combined")
	}
	if !fromStdin {
		return value, valueSet, nil
	}
	endpoint, err := utils.ResolveEndpoint(resource)
	if err != nil || endpoint != "secrets" {
		return "", false, fmt.Errorf("--value-stdin is only supported for secrets")
	}
	const maxSecretSize = 1 << 20
	data, err := io.ReadAll(io.LimitReader(cmd.InOrStdin(), maxSecretSize+1))
	if err != nil {
		return "", false, fmt.Errorf("read secret from stdin: %w", err)
	}
	if len(data) > maxSecretSize {
		return "", false, fmt.Errorf("secret exceeds %d bytes", maxSecretSize)
	}
	return string(data), true, nil
}

func init() {
	pkg.RegisterCommand(createCmd)
}
