// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"dhcli/keys"

	"github.com/spf13/viper"
	"gopkg.in/ini.v1"
)

// GetConfigEntries returns key-value pairs for all configuration fields:
// keys present in the current INI section that are not credentials and
// not internal CLI keys. Values are read from Viper (env may override).
func GetConfigEntries() map[string]string {
	result := make(map[string]string)

	credSet := credentialKeySet()

	cfg, err := ini.Load(getIniPath())
	if err != nil {
		return result
	}

	env := viper.GetString(keys.CurrentEnvironment)
	if env == "" {
		env = "default"
	}

	for _, k := range cfg.Section(env).Keys() {
		name := k.Name()
		if credSet[name] || internalKeys[name] || name == keys.CredentialsList {
			continue
		}
		result[name] = viper.GetString(name)
	}
	return result
}

// GetCredentialEntries returns key-value pairs for all credential fields,
// as recorded in credentials_list by the most recent login or refresh.
// Returns an empty map if credentials_list is absent (e.g. old INI).
func GetCredentialEntries() map[string]string {
	result := make(map[string]string)
	for _, k := range splitCSV(viper.GetString(keys.CredentialsList)) {
		result[k] = viper.GetString(k)
	}
	return result
}

// credentialKeySet returns the set of keys listed in credentials_list.
func credentialKeySet() map[string]bool {
	set := make(map[string]bool)
	for _, k := range splitCSV(viper.GetString(keys.CredentialsList)) {
		set[k] = true
	}
	return set
}
