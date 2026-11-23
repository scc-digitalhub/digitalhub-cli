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
		// Set logger mode and debug/trace mode based on flags
		verboseFlag := cmd.Flags().Lookup("verbose")
		debugFlag := cmd.Flags().Lookup("debug")
		traceFlag := cmd.Flags().Lookup("trace")

		isDebug := debugFlag != nil && debugFlag.Value.String() == "true"
		isTrace := traceFlag != nil && traceFlag.Value.String() == "true"

		// Debug or trace flag enables verbose mode automatically
		isVerbose := (verboseFlag != nil && verboseFlag.Value.String() == "true") || isDebug || isTrace

		if isVerbose {
			utils.GetGlobalLogger().SetMode(utils.ModeVerbose)
		} else {
			utils.GetGlobalLogger().SetMode(utils.ModeQuiet)
		}

		// Enable HTTP logging transport if trace or debug flag is set
		if isTrace {
			utils.EnableTraceMode()
		} else if isDebug {
			utils.EnableDebugMode()
		}

		envFlag := cmd.Flags().Lookup("env")
		var env string
		if envFlag != nil && envFlag.Value.Type() == "string" && envFlag.Value.String() != "" {
			env = envFlag.Value.String()
		}
		if environmentFlag := cmd.Flags().Lookup("environment"); environmentFlag != nil {
			env = environmentFlag.Value.String()
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
	// Add persistent verbose, debug, and trace flags to root command
	dhcli.PersistentFlags().BoolP("verbose", "v", false, "enable verbose output")
	dhcli.PersistentFlags().Bool("debug", false, "enable HTTP debug logging")
	dhcli.PersistentFlags().Bool("trace", false, "enable full HTTP trace logging (includes headers, query parameters, and bodies)")
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
