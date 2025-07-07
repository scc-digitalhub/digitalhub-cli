// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"context"
	"fmt"
	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
	"os"
)

var dhcli = &cobra.Command{
	Use:   "dhcli",
	Short: "dhcli is a tool for managing resource in core platform",
	Long:  `dhcli is a command-line utility for downloading, uploading, and managing core platform entity`,
}

func Execute() {

	if err := fang.Execute(context.Background(), dhcli); err != nil {
		_, err := fmt.Fprintln(os.Stderr, err)
		if err != nil {
			return
		}
		os.Exit(1)
	}

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
