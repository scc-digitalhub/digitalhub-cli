// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"dhcli/handlers/adapter"
	"dhcli/pkg"

	"github.com/spf13/cobra"
)

func newExtensionCommand() *cobra.Command {
	var project, file, output string
	var resetID bool
	cmd := &cobra.Command{
		Use:   "extension <create|update|delete|get|list> <parent-resource> <parent-id> [id]",
		Short: "Manage extensions of a project or resource",
		Long: `Manage extensions of a project or resource.
Provide the parent resource and parent id. For non-project parents, -p is
required. For get, update, and delete, the final id is the extension's id.`,
		Args: func(cmd *cobra.Command, args []string) error {
			opts, err := parseExtensionArgs(args)
			if err != nil {
				return err
			}
			if opts.Parent != "project" && opts.Parent != "projects" && project == "" {
				return fmt.Errorf("--project is required for %s extensions", opts.Parent)
			}
			if (args[0] == "create" || args[0] == "update") && file == "" {
				return fmt.Errorf("%s requires --file", args[0])
			}
			if args[0] != "create" && args[0] != "update" && file != "" {
				return fmt.Errorf("--file is only valid for create and update")
			}
			if resetID && args[0] != "create" {
				return fmt.Errorf("--reset-id is only valid for create")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := parseExtensionArgs(args)
			if err != nil {
				return err
			}
			opts.Project = project
			opts.File, opts.Output, opts.ResetID = file, output, resetID
			return adapter.ExtensionHandler(cmd.Context(), cmd.OutOrStdout(), opts)
		},
	}
	cmd.Flags().StringVarP(&project, "project", "p", "", "project (required for non-project parents)")
	cmd.Flags().StringVarP(&file, "file", "f", "", "YAML extension definition for create or update")
	cmd.Flags().StringVarP(&output, "out", "o", "yaml", "output format for get or list (short, json, yaml)")
	cmd.Flags().BoolVarP(&resetID, "reset-id", "r", false, "remove the ID from the file when creating an extension")
	return cmd
}

func parseExtensionArgs(args []string) (adapter.ExtensionOptions, error) {
	var opts adapter.ExtensionOptions
	if len(args) == 0 {
		return opts, fmt.Errorf("an action is required")
	}
	opts.Action = args[0]
	args = args[1:]
	needsID := false
	switch opts.Action {
	case "create", "list":
	case "update", "delete", "get":
		needsID = true
	default:
		return opts, fmt.Errorf("unknown extension action %q", opts.Action)
	}
	if needsID {
		if len(args) == 0 {
			return opts, fmt.Errorf("%s requires an extension ID", opts.Action)
		}
		opts.ID = args[len(args)-1]
		args = args[:len(args)-1]
	}
	if len(args) != 2 {
		return opts, fmt.Errorf("expected <parent-resource> <parent-id>")
	}
	opts.Parent, opts.ParentID = args[0], args[1]
	if opts.Parent == "" || opts.ParentID == "" || (needsID && opts.ID == "") {
		return opts, fmt.Errorf("positional arguments must not be empty")
	}
	return opts, nil
}

func init() {
	pkg.RegisterCommand(newExtensionCommand())
}
