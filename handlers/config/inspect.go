// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"dhcli/handlers/utils"
	"sigs.k8s.io/yaml"
)

func ConfigHandler(output string) error {
	entries := utils.GetConfigEntries()
	return printEntries(entries, utils.TranslateFormat(output))
}

func CredentialsHandler(output string) error {
	entries := utils.GetCredentialEntries()
	return printEntries(entries, utils.TranslateFormat(output))
}

func printEntries(entries map[string]string, format string) error {
	if len(entries) == 0 {
		fmt.Println("No entries found.")
		return nil
	}

	switch format {
	case "json":
		b, err := json.MarshalIndent(entries, "", "    ")
		if err != nil {
			return err
		}
		fmt.Println(string(b))
	case "yaml":
		b, err := yaml.Marshal(entries)
		if err != nil {
			return err
		}
		fmt.Print(string(b))
	default:
		keys := make([]string, 0, len(entries))
		for k := range entries {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Printf("%s=%s\n", strings.ToUpper(k), entries[k])
		}
	}

	return nil
}
