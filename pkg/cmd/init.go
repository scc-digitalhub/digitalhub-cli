// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"dhcli/handlers/environment"
	"dhcli/handlers/utils"
	"dhcli/pkg"
	"dhcli/pkg/flags"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var initCmd = func() *cobra.Command {
	// Local flag declarations
	envFlag := flags.NewStringFlag("env", "e", "environment", "")
	preFlag := flags.NewBoolFlag("pre", "", "Include pre-release versions when installing", false)

	cmd := &cobra.Command{
		Use:   "init [path]",
		Short: "Initialize a Python virtual environment and install packages",
		Long:  "Creates or activates a Python virtual environment at the specified path (or current environment name) and installs dependencies. If no path is provided, uses the current environment name. If the path already contains a valid venv, dependencies are installed without recreating it.",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Load configuration to access current environment
			if err := utils.RegisterIniCfgWithViper(*envFlag.Value); err != nil {
				log.Fatalf("Failed to load configuration: %v", err)
			}

			venvPath := ""
			if len(args) > 0 {
				venvPath = args[0]
			} else {
				// If no path provided, use current environment name
				venvPath = viper.GetString(utils.CurrentEnvironment)
			}

			// If still empty, default to current directory
			if venvPath == "" {
				venvPath = "."
			}

			if err := environment.InitEnvironmentHandler(venvPath, *preFlag.Value); err != nil {
				log.Fatalf("Init failed: %v", err)
			}
		},
	}

	flags.AddFlag(cmd, &envFlag)
	flags.AddFlag(cmd, &preFlag)

	return cmd
}()

func init() {
	pkg.RegisterCommand(initCmd)
}
