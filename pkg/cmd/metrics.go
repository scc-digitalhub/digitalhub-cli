// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"dhcli/handlers/adapter"
	"dhcli/pkg"
	"dhcli/pkg/flags"
	"log"

	"dhcli/handlers/utils"

	"github.com/spf13/cobra"
)

var metricsCmd = func() *cobra.Command {
	envFlag := flags.NewStringFlag("env", "e", "environment", "")
	outFlag := flags.NewStringFlag("out", "o", "output format (short, json, yaml)", "")
	followFlag := flags.NewBoolFlag("follow", "f", "Continuously refresh metrics every 15 seconds", false)

	cmd := &cobra.Command{
		Use:   "metrics",
		Short: "Read resource metrics",
		Long:  "Read resource metrics. Use subcommands 'project' or 'run' for scoped metrics; omit for instance-wide metrics.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			err := adapter.MetricsHandler(
				*envFlag.Value,
				*outFlag.Value,
				"",
				"instance",
				"",
				*followFlag.Value,
			)
			if err != nil {
				log.Fatalf("Failed: %v", err)
			}
		},
	}

	flags.AddFlag(cmd, &envFlag)
	flags.AddFlag(cmd, &outFlag)
	flags.AddFlag(cmd, &followFlag)

	return cmd
}()

var metricsProjectCmd = func() *cobra.Command {
	envFlag := flags.NewStringFlag("env", "e", "environment", "")
	outFlag := flags.NewStringFlag("out", "o", "output format (short, json, yaml)", "")
	followFlag := flags.NewBoolFlag("follow", "f", "Continuously refresh metrics every 15 seconds", false)

	cmd := &cobra.Command{
		Use:   "project <name>",
		Short: "Read project resource metrics",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := adapter.MetricsHandler(
				*envFlag.Value,
				*outFlag.Value,
				args[0],
				"project",
				"",
				*followFlag.Value,
			)
			if err != nil {
				log.Fatalf("Failed: %v", err)
			}
		},
	}

	flags.AddFlag(cmd, &envFlag)
	flags.AddFlag(cmd, &outFlag)
	flags.AddFlag(cmd, &followFlag)

	return cmd
}()

var metricsRunCmd = func() *cobra.Command {
	envFlag := flags.NewStringFlag("env", "e", "environment", "")
	outFlag := flags.NewStringFlag("out", "o", "output format (short, json, yaml)", "")
	projectFlag := flags.NewStringFlag("project", "p", "Project name (required)", "")
	followFlag := flags.NewBoolFlag("follow", "f", "Continuously refresh metrics every 15 seconds", false)

	cmd := &cobra.Command{
		Use:   "run <id>",
		Short: "Read run resource metrics",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			project := utils.ResolveProject(*projectFlag.Value)
			err := adapter.MetricsHandler(
				*envFlag.Value,
				*outFlag.Value,
				project,
				"run",
				args[0],
				*followFlag.Value,
			)
			if err != nil {
				log.Fatalf("Failed: %v", err)
			}
		},
	}

	flags.AddFlag(cmd, &envFlag)
	flags.AddFlag(cmd, &outFlag)
	flags.AddFlag(cmd, &projectFlag)
	flags.AddFlag(cmd, &followFlag)

	return cmd
}()

func init() {
	metricsCmd.AddCommand(metricsProjectCmd)
	metricsCmd.AddCommand(metricsRunCmd)
	pkg.RegisterCommand(metricsCmd)
}
