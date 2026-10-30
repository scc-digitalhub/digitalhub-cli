// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"log"

	"dhcli/handlers/config"
	"dhcli/handlers/environment"
	"dhcli/handlers/utils"
	"dhcli/keys"
	"dhcli/pkg"
	"dhcli/pkg/flags"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var initCmd = func() *cobra.Command {
	// Local flag declarations
	environmentFlag := flags.NewStringFlag("environment", "e", "Registered environment to use", "")
	envFlag := flags.NewBoolFlag("env", "", "Write configuration and credentials to a local .env file without creating a Python venv", false)
	preFlag := flags.NewBoolFlag("pre", "", "Include pre-release versions when installing", false)
	agentFlag := flags.NewBoolFlag("agent", "", "Install embedded agent templates with knowledge links matching the connected Core version", false)

	cmd := &cobra.Command{
		Use:   "init [path]",
		Short: "Initialize a Python virtual environment, write a .env file or install agent templates",
		Long:  "Creates or activates a Python virtual environment at the specified path (or current environment name) and installs dependencies. If no path is provided, uses the current environment name. If the path already contains a valid venv, dependencies are installed without recreating it. With --agent, fetches the current Core version and installs embedded AGENTS.md, CLAUDE.md and .agents/skills templates in the current working directory (without creating a Python venv or installing SDK packages), with remote knowledge links pinned to the matching major.minor docs branch. Identical agent files are left unchanged; conflicting files are never overwritten. With --env, merges configuration and refreshed credentials into .env in the current working directory without creating a Python venv. Existing variables are preserved unless replaced by exported values; formatting and comments are normalized. The file is atomically replaced with owner-only permissions. Use --environment NAME or -e NAME to select a registered environment.",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Load configuration to access current environment
			if err := utils.RegisterIniCfgWithViper(*environmentFlag.Value); err != nil {
				log.Fatalf("Failed to load configuration: %v", err)
			}

			if *envFlag.Value {
				if len(args) > 0 {
					log.Fatalf("Init failed: --env does not accept a [path] argument (.env is written in the current working directory)")
				}
				if *agentFlag.Value || *preFlag.Value {
					log.Fatalf("Init failed: --env cannot be combined with --agent or --pre")
				}
				if err := config.EnvFileHandler(".env"); err != nil {
					log.Fatalf("Init failed: %v", err)
				}
				log.Println("Configuration and credentials written to .env (owner-only permissions). Keep this file out of version control.")
				return
			}

			if *agentFlag.Value {
				if len(args) > 0 {
					log.Fatalf("Init failed: --agent does not accept a [path] argument (agent templates are installed in the current working directory)")
				}
				if *preFlag.Value {
					log.Fatalf("Init failed: --pre cannot be combined with --agent")
				}
				if err := environment.InitAgentWorkspaceHandler(); err != nil {
					log.Fatalf("Init failed: %v", err)
				}
				return
			}

			venvPath := ""
			if len(args) > 0 {
				venvPath = args[0]
			} else {
				// If no path provided, use current environment name
				venvPath = viper.GetString(keys.CurrentEnvironment)
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

	flags.AddFlag(cmd, &environmentFlag)
	flags.AddFlag(cmd, &envFlag)
	flags.AddFlag(cmd, &preFlag)
	flags.AddFlag(cmd, &agentFlag)

	return cmd
}()

func init() {
	pkg.RegisterCommand(initCmd)
}
