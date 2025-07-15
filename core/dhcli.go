// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"context"
	"dhcli/utils"
	"fmt"
	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var dhcli = &cobra.Command{
	Use:   "dhcli",
	Short: "dhcli is a tool for managing resource in core platform",
	Long:  `dhcli is a command-line utility for downloading, uploading, and managing core platform entity`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Parse --env if present
		envFlag := cmd.Flags().Lookup("env")
		var env string
		if envFlag != nil && envFlag.Value.String() != "" {
			env = envFlag.Value.String()
		}

		// Reload ini config with a correct section
		utils.RegisterIniCfgWithViper(env)

		//// Bind all flags after loading config
		//utils.BindFlagsToViperRecursive(cmd.Root())
		return nil
	},
}

func Execute() {

	// Show final config
	fmt.Println("📦 Final Config:")
	for _, key := range viper.AllKeys() {
		fmt.Printf("%s = %v\n", key, viper.Get(key))
	}

	if err := fang.Execute(context.Background(), dhcli); err != nil {
		_, err := fmt.Fprintln(os.Stderr, err)
		if err != nil {
			return
		}
		os.Exit(1)
	}

	// Uncomment this and commet the code above to use the original cobra.Execute() method
	//if err := dhcli.Execute(); err != nil {
	//	_, err := fmt.Fprintln(os.Stderr, err)
	//	if err != nil {
	//		return
	//	}
	//	os.Exit(1)
	//}
}

func RegisterCommand(cmd *cobra.Command) {
	dhcli.AddCommand(cmd)
}
