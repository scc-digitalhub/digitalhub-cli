// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package pkg

import (
	"context"
	"fmt"
	"os"
	"slices"

	"dhcli/handlers/utils"

	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
)

var version string

var dhcli = &cobra.Command{
	Use:   "dhcli",
	Short: "dhcli is a tool for managing resource in core platform",
	Long:  `dhcli is a command-line utility for downloading, uploading, and managing core platform entity`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Set logger mode and debug mode based on flags
		verboseFlag := cmd.Flags().Lookup("verbose")
		debugFlag := cmd.Flags().Lookup("debug")

		// Debug flag enables verbose mode automatically
		isVerbose := (verboseFlag != nil && verboseFlag.Value.String() == "true") ||
			(debugFlag != nil && debugFlag.Value.String() == "true")

		if isVerbose {
			utils.GetGlobalLogger().SetMode(utils.ModeVerbose)
		} else {
			utils.GetGlobalLogger().SetMode(utils.ModeQuiet)
		}

		// Enable debug HTTP transport if flag is set
		if debugFlag != nil && debugFlag.Value.String() == "true" {
			utils.EnableDebugMode()
		}

		envFlag := cmd.Flags().Lookup("env")
		var env string
		if envFlag != nil && envFlag.Value.String() != "" {
			env = envFlag.Value.String()
		}

		// Only skip config for explicit maintenance cmds
		if !(slices.Contains([]string{"register", "use", "remove", "list-env"}, cmd.Name())) {
			if err := utils.RegisterIniCfgWithViper(env); err != nil {
				return err
			}
		}

		// Show final config
		//fmt.Println("📦 Final Config:")
		//for _, key := range viper.AllKeys() {
		//	fmt.Printf("%s = %v\n", key, viper.Get(key))
		//}

		return nil
	},
}

func init() {
	// Add persistent verbose flag to root command
	dhcli.PersistentFlags().BoolP("verbose", "v", false, "enable verbose output")
	dhcli.PersistentFlags().Bool("debug", false, "enable HTTP debug logging")
}

func Execute() {
	var opts []fang.Option
	if version != "" {
		opts = append(opts, fang.WithVersion(version))
	}
	if err := fang.Execute(context.Background(), dhcli, opts...); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func RegisterCommand(cmd *cobra.Command) {
	dhcli.AddCommand(cmd)
}
