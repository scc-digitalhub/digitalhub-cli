// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package flags

import (
	"github.com/spf13/cobra"
)

type SpecificCommandFlag struct {
	OutputFlag   string
	CreateFlag   bool
	InputFlag    string
	FilePathFlag string
	ResetIdFlag  bool
	ConfirmFlag  bool
	CascadeFlag  bool
	PreFlag      bool
	ListKind     string
	ListState    string
	EnvFlag      string
	IdFlag       string
}

type commonCommandFlag struct {
	EnvFlag     string
	OutFlag     string
	ProjectFlag string
	NameFlag    string
}

var CommonFlag = commonCommandFlag{}

func AddCommonFlags(cmd *cobra.Command, flagsToAdd ...string) {

	if len(flagsToAdd) == 0 {
		flagsToAdd = []string{"env", "out", "project", "name"}
	}

	for _, flag := range flagsToAdd {
		switch flag {
		case "env":
			cmd.Flags().StringVarP(&CommonFlag.EnvFlag, "env", "e", "", "environment")
		case "out":
			cmd.Flags().StringVarP(&CommonFlag.OutFlag, "out", "o", "short", "output format (short, json, yaml)")
		case "project":
			cmd.Flags().StringVarP(&CommonFlag.ProjectFlag, "project", "p", "", "project")
		case "name":
			cmd.Flags().StringVarP(&CommonFlag.NameFlag, "name", "n", "", "name")
		}
	}
}
