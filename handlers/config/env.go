// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"dhcli/handlers/utils"

	"github.com/subosito/gotenv"
)

var envNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func EnvFileHandler(path string) error {
	utils.CheckUpdateEnvironment()
	if err := utils.CheckCredentials(); err != nil {
		return err
	}
	entries := getConfigEntriesByProvider("", "")
	for name, value := range getCredentialEntriesByProvider("", "") {
		entries[name] = value
	}
	if len(entries) == 0 {
		return fmt.Errorf("no configuration or credentials found; register and select an environment first")
	}
	return writeEnvFile(path, entries)
}

func writeEnvFile(path string, entries map[string]string) error {
	merged := gotenv.Env{}
	info, err := os.Lstat(path)
	if err == nil {
		if !info.Mode().IsRegular() {
			return fmt.Errorf("refusing to replace non-regular env file %s", path)
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read env file: %w", err)
		}
		merged, err = gotenv.StrictParse(strings.NewReader(string(content)))
		if err != nil {
			return fmt.Errorf("cannot parse %s as dotenv; file left unchanged", path)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("inspect env file: %w", err)
	}
	for name, value := range entries {
		merged[strings.ToUpper(name)] = value
	}
	names := make([]string, 0, len(merged))
	for name, value := range merged {
		if !envNamePattern.MatchString(name) {
			return fmt.Errorf("configuration contains an invalid shell variable name")
		}
		if strings.ContainsRune(value, 0) {
			return fmt.Errorf("configuration contains a NUL byte in %s", name)
		}
		names = append(names, name)
	}
	sort.Strings(names)
	var content strings.Builder
	escape := strings.NewReplacer(`\`, `\\`, `"`, `\"`, `$`, `\$`, "`", "\\`", "\n", `\n`, "\r", `\r`)
	for _, name := range names {
		fmt.Fprintf(&content, "%s=\"%s\"\n", name, escape.Replace(merged[name]))
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".dhcli-env-*")
	if err != nil {
		return fmt.Errorf("create private env file: %w", err)
	}
	defer os.Remove(file.Name())
	if _, err := file.WriteString(content.String()); err != nil {
		file.Close()
		return fmt.Errorf("write env file: %w", err)
	}
	if err := file.Sync(); err != nil {
		file.Close()
		return fmt.Errorf("sync env file: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close env file: %w", err)
	}
	if err := os.Rename(file.Name(), path); err != nil {
		return fmt.Errorf("replace env file: %w", err)
	}
	return nil
}
