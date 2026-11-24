// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"log"

	"dhcli/handlers/config"
	"dhcli/handlers/environment"
	"dhcli/keys"
	"dhcli/pkg"
	"dhcli/pkg/flags"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func initCommand() *cobra.Command {
	// Local flag declarations
	environmentFlag := flags.NewStringFlag("environment", "", "Registered environment to use (overrides --env)", "")
	venvFlag := flags.NewBoolFlag("venv", "", "Initialize a Python virtual environment", false)
	envFileFlag := flags.NewBoolFlag("env-file", "", "Write configuration and credentials to a local .env file", false)
	preFlag := flags.NewBoolFlag("pre", "", "Include pre-release versions when installing", false)
	agentFlag := flags.NewBoolFlag("agent", "", "Install latest agent templates (or embedded fallback) with knowledge links matching the connected Core version", false)
	configFlag := flags.NewBoolFlag("config", "", "Write the current environment to a local .dhcore.ini file", false)
	overwriteFlag := flags.NewBoolFlag("overwrite", "", "Update existing venv and initialization files without prompting", false)
	skipGitignoreFlag := flags.NewBoolFlag("skip-gitignore", "", "Do not create or update .gitignore", false)

	cmd := &cobra.Command{
		Use:   "init [path]",
		Short: "Initialize a virtual environment, .env, agent templates and local configuration",
		Long:  "By default, initializes a Python virtual environment, writes a .env file, installs agent templates and writes a local .dhcore.ini file, in that order. Use any combination of --venv, --env-file, --agent and --config to run only the selected steps. Finally, .gitignore is created or updated for the outputs of the selected steps; use --skip-gitignore to opt out. Paths outside the current directory cannot be ignored by this file. Git ignore rules do not untrack files already tracked by Git. The optional path applies to the venv and defaults to .venv. An existing valid venv is reused after confirmation, or updated without prompting with --overwrite. --overwrite upgrades DigitalHub packages in place without recreating the venv; dependencies may also change. --pre includes pre-release dependencies when initializing a venv. The .env, .dhcore.ini and agent templates are written in the current working directory. The local .dhcore.ini contains only the selected environment and is replaced entirely with --overwrite. DH_CONFIG takes precedence over the local file. Agent templates are fetched from the CLI repository (falling back to embedded templates if offline), with knowledge links pinned to the connected Core's major.minor docs branch. Existing differing files are not overwritten unless --overwrite is passed; embedded agent templates never overwrite existing files. An existing .env is merged when overwritten, preserving custom variables unless replaced by exported values; formatting and comments are normalized. Both configuration files are written with owner-only permissions and may contain credentials. Use --env NAME or -e NAME to select a registered environment.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			venv, envFile, agent, localConfig := *venvFlag.Value, *envFileFlag.Value, *agentFlag.Value, *configFlag.Value
			selectionSpecified := cmd.Flags().Changed("venv") || cmd.Flags().Changed("env-file") || cmd.Flags().Changed("agent") || cmd.Flags().Changed("config")
			if !selectionSpecified {
				venv, envFile, agent, localConfig = true, true, true, true
			}
			if !venv && !envFile && !agent && !localConfig {
				return fmt.Errorf("init failed: no initialization steps selected")
			}

			if !venv && len(args) > 0 {
				return fmt.Errorf("init failed: [path] requires the venv step")
			}
			if !venv && *preFlag.Value {
				return fmt.Errorf("init failed: --pre requires the venv step")
			}
			var paths []string
			venvPath := ".venv"
			if venv {
				if len(args) > 0 {
					venvPath = args[0]
				}
				if venvPath == "" {
					venvPath = "."
				}
				if err := environment.InitEnvironmentHandler(venvPath, viper.GetString(keys.CurrentEnvironment), *preFlag.Value, *overwriteFlag.Value); err != nil {
					return fmt.Errorf("init failed: %w", err)
				}
				paths = append(paths, venvPath)
			}
			if envFile {
				if err := config.EnvFileHandler(".env", *overwriteFlag.Value); err != nil {
					return fmt.Errorf("init failed: %w", err)
				}
				log.Println("Configuration and credentials written to .env (owner-only permissions). Keep this file out of version control.")
				paths = append(paths, ".env")
			}
			if agent {
				if err := environment.InitAgentWorkspaceHandler(*overwriteFlag.Value); err != nil {
					return fmt.Errorf("init failed: %w", err)
				}
				paths = append(paths, "AGENTS.md", "CLAUDE.md", ".agents/skills")
			}
			if localConfig {
				if err := config.LocalConfigHandler(keys.IniName, *overwriteFlag.Value); err != nil {
					return fmt.Errorf("init failed: %w", err)
				}
				log.Println("Current environment written to .dhcore.ini (owner-only permissions). Keep this file out of version control.")
				paths = append(paths, keys.IniName)
			}
			if !*skipGitignoreFlag.Value {
				if err := environment.UpdateInitGitignore(paths); err != nil {
					return fmt.Errorf("init failed: %w", err)
				}
			}
			return nil
		},
	}

	flags.AddFlag(cmd, &environmentFlag)
	flags.AddFlag(cmd, &venvFlag)
	flags.AddFlag(cmd, &envFileFlag)
	flags.AddFlag(cmd, &preFlag)
	flags.AddFlag(cmd, &agentFlag)
	flags.AddFlag(cmd, &configFlag)
	flags.AddFlag(cmd, &overwriteFlag)
	flags.AddFlag(cmd, &skipGitignoreFlag)

	return cmd
}

func init() {
	pkg.RegisterCommand(initCommand())
}
