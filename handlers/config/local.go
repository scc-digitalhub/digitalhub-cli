// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"os"
	"path/filepath"

	"dhcli/handlers/utils"
	"dhcli/keys"

	"github.com/spf13/viper"
)

func LocalConfigHandler(path string, overwrite bool) error {
	env := viper.GetString(keys.CurrentEnvironment)
	if env == "" {
		return fmt.Errorf("no current environment selected")
	}
	if viper.GetString(keys.DhCoreEndpoint) == "" {
		return fmt.Errorf("current environment %q has no Core endpoint", env)
	}

	info, err := os.Lstat(path)
	if err == nil {
		if !info.Mode().IsRegular() {
			return fmt.Errorf("refusing to replace non-regular config file %s", path)
		}
		if !overwrite {
			return fmt.Errorf("refusing to overwrite existing config file %s; pass --overwrite to replace it", path)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("inspect config file: %w", err)
	}

	file, err := os.CreateTemp(filepath.Dir(path), ".dhcli-config-*")
	if err != nil {
		return fmt.Errorf("create private config file: %w", err)
	}
	defer os.Remove(file.Name())
	if err := file.Close(); err != nil {
		return fmt.Errorf("close config file: %w", err)
	}

	var configKeys []string
	for _, key := range viper.AllKeys() {
		if key != keys.CurrentEnvironment && !utils.InternalKeys[key] {
			configKeys = append(configKeys, key)
		}
	}
	if err := utils.PersistToIni(file.Name(), env, configKeys); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}
	if overwrite {
		if err := os.Rename(file.Name(), path); err != nil {
			return fmt.Errorf("replace config file: %w", err)
		}
	} else if err := os.Link(file.Name(), path); err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("refusing to overwrite existing config file %s; pass --overwrite to replace it", path)
		}
		return fmt.Errorf("publish config file: %w", err)
	}
	return nil
}
