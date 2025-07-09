// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
	"time"

	"gopkg.in/ini.v1"
)

//func CheckUpdateEnvironment(cfg *ini.File, section *ini.Section) {
//	if section.HasKey(UpdatedEnvKey) {
//		updated, err := time.Parse(time.RFC3339, section.Key(UpdatedEnvKey).Value())
//		if err != nil || updated.Add(outdatedAfterHours*time.Hour).Before(time.Now()) {
//			updateEnvironment(cfg, section)
//		}
//	}
//}

//func updateEnvironment(cfg *ini.File, section *ini.Section) {
//	baseEndpoint := section.Key(DhCoreEndpoint).Value()
//	if baseEndpoint == "" {
//		return
//	}
//
//	// Configuration
//	config, err := FetchConfig(baseEndpoint + "/.well-known/configuration")
//	if err != nil {
//		return
//	}
//	for k, v := range config {
//		newKey := k
//		if newKey == ClientIdKey {
//			newKey = "client_id"
//		}
//		UpdateKey(section, newKey, v)
//	}
//
//	// OpenID Configuration
//	openIdConfig, err := FetchConfig(baseEndpoint + "/.well-known/openid-configuration")
//	if err != nil {
//		return
//	}
//	for _, k := range OpenIdFields {
//		if v, ok := openIdConfig[k]; ok && v != "" {
//			UpdateKey(section, k, v)
//		}
//	}
//
//	// Update timestamp
//	section.Key(UpdatedEnvKey).SetValue(time.Now().Format(time.RFC3339))
//	SaveIni(cfg)
//}

func UpdateKey(section *ini.Section, k string, v interface{}) {
	if !section.HasKey(k) {
		section.NewKey(k, ReflectValue(v))
	} else {
		section.Key(k).SetValue(ReflectValue(v))
	}
}

func CheckUpdateEnvironment() {
	if viper.Get(UpdatedEnvKey) != nil {
		updated, err := time.Parse(time.RFC3339, viper.GetString(UpdatedEnvKey))
		if err != nil || updated.Add(outdatedAfterHours*time.Hour).Before(time.Now()) {
			updateEnvironment()
		}
		return

	}
}

func updateEnvironment() {
	baseEndpoint := viper.GetString(DhCoreEndpoint)
	if baseEndpoint == "" {
		return
	}

	// Configuration
	config, err := FetchConfig(baseEndpoint + "/.well-known/configuration")
	if err != nil {
		return
	}
	for k, v := range config {
		newKey := k
		if newKey == ClientIdKey {
			newKey = "client_id"
		}
		// Update the key in the viper config
		viper.Set(newKey, v)
		//UpdateKey(section, newKey, v)
	}

	// OpenID Configuration
	openIdConfig, err := FetchConfig(baseEndpoint + "/.well-known/openid-configuration")
	if err != nil {
		return
	}
	for _, k := range OpenIdFields {
		if v, ok := openIdConfig[k]; ok && v != "" {
			viper.Set(k, v)
			//UpdateKey(section, k, v)
		}
	}

	// Update timestamp
	viper.Set(UpdatedEnvKey, time.Now().Format(time.RFC3339))
	//section.Key(UpdatedEnvKey).SetValue(time.Now().Format(time.RFC3339))
	err = UpdateIniSectionFromViper(viper.AllKeys())
	if err != nil {
		return
	}
}

func UpdateIniSectionFromViper(keys []string) error {

	//print all keys to update viper.allKeys() foreach
	// This function updates the INI file section with keys and values from Viper.
	for _, key := range viper.AllKeys() {
		fmt.Printf("Enum key %s\n", key)
	}

	iniPath := os.ExpandEnv("$HOME/" + IniName)

	cfg, err := ini.Load(iniPath)
	if err != nil {
		return fmt.Errorf("failed to load ini file: %w", err)
	}

	// Get the currently selected environment
	env := viper.GetString("current_environment")
	if env == "" {
		env = "DEFAULT"
	}

	section := cfg.Section(env)

	// Update keys from Viper back into the ini section
	for _, key := range keys {
		val := viper.GetString(key)
		if key == "current_environment" {
			// Skip updating the current_environment key in the ini file
			continue
		}
		section.Key(key).SetValue(val)
	}

	// Save updated INI
	err = cfg.SaveTo(iniPath)
	if err != nil {
		return fmt.Errorf("failed to save ini file: %w", err)
	}

	fmt.Printf("💾 Updated section [%s] in %s\n", env, iniPath)
	return nil
}
