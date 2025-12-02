// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"context"
	"dhcli/sdk/utils"
	"fmt"
	"os"
	"slices"

	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
)

var dhcli = &cobra.Command{
	Use:   "dhcli",
	Short: "dhcli is a tool for managing resource in core platform",
	Long:  `dhcli is a command-line utility for downloading, uploading, and managing core platform entity`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
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

func Execute() {
	if err := fang.Execute(context.Background(), dhcli); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func RegisterCommand(cmd *cobra.Command) {
	dhcli.AddCommand(cmd)
}
