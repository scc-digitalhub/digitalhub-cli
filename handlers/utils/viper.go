// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"time"

	"dhcli/keys"

	"github.com/spf13/viper"
	"gopkg.in/ini.v1"
)

// internalKeys are CLI-only keys that must never be added as new INI entries.
var internalKeys = map[string]bool{}

// SetupViperEnv configures Viper to automatically bind environment variables.
// Key foo_bar maps to env var FOO_BAR.
func SetupViperEnv() {
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
}

// PersistToIni updates every existing key in the named INI section from the
// current Viper value, then upserts any explicitly-provided additional keys.
// All values are written as-is, including empty strings.
// If the INI file does not yet exist a new one is created.
func PersistToIni(iniPath, envName string, additionalKeys []string) error {
	cfg, err := ini.Load(iniPath)
	if err != nil {
		cfg = ini.Empty()
		cfg.Section("DEFAULT").Key(keys.CurrentEnvironment).SetValue(envName)
	}

	sec := cfg.Section(envName)

	// Update all existing section keys from Viper.
	for _, k := range sec.Keys() {
		name := k.Name()
		if internalKeys[name] {
			continue
		}
		k.SetValue(viper.GetString(name))
	}

	// Upsert explicitly-provided additional keys in sorted order.
	sort.Strings(additionalKeys)
	for _, name := range additionalKeys {
		if internalKeys[name] {
			continue
		}
		if sec.HasKey(name) {
			sec.Key(name).SetValue(viper.GetString(name))
		} else {
			sec.NewKey(name, viper.GetString(name))
		}
	}

	if !cfg.Section("DEFAULT").HasKey(keys.CurrentEnvironment) {
		cfg.Section("DEFAULT").Key(keys.CurrentEnvironment).SetValue(envName)
	}
	return cfg.SaveTo(iniPath)
}

// PersistCurrentEnv is a convenience wrapper around PersistToIni that resolves
// the INI path and current environment name from Viper.
func PersistCurrentEnv(additionalKeys []string) error {
	env := viper.GetString(keys.CurrentEnvironment)
	if env == "" {
		env = resolveEnvName()
	}
	return PersistToIni(getIniPath(), env, additionalKeys)
}

// resolveEnvName: --env > "default"
func resolveEnvName(optionalEnv ...string) string {
	if len(optionalEnv) > 0 && optionalEnv[0] != "" && strings.ToLower(optionalEnv[0]) != "null" {
		return optionalEnv[0]
	}
	return "default"
}

// Load [DEFAULT] + [env] into Viper (TOML in-memory). ENV can still override on Get().
func loadIniSectionIntoViper(cfg *ini.File, env string) error {
	def := cfg.Section("DEFAULT")
	selected := def
	if env != "" && cfg.HasSection(env) {
		selected = cfg.Section(env)
		logger.Info(fmt.Sprintf("Using env: [%s]", env))
	} else if env == "" || strings.EqualFold(env, "DEFAULT") {
		logger.Info("Using env: [DEFAULT]")
	} else {
		logger.Warn("Env not found, falling back to [DEFAULT]")
	}

	merged := make(map[string]string)
	for _, k := range def.Keys() {
		merged[k.Name()] = k.Value()
	}
	if selected != nil && selected != def {
		for _, k := range selected.Keys() {
			merged[k.Name()] = k.Value()
		}
	}

	var buf bytes.Buffer
	for k, v := range merged {
		vSafe := strings.ReplaceAll(strings.ReplaceAll(v, `\`, `\\`), `"`, `\"`)
		_, _ = fmt.Fprintf(&buf, "%s = \"%s\"\n", k, vSafe)
	}
	viper.SetConfigType("toml")
	return viper.ReadConfig(&buf)
}

// RegisterIniCfgWithViper:
// 1) bind ENV from struct (live)
// 2) load INI or lazy-bootstraps it from well-known (writes only target env)
// 3) load active section into Viper and set current_environment
func RegisterIniCfgWithViper(optionalEnv ...string) error {
	iniPath := getIniPath()

	SetupViperEnv()

	cfg, err := ini.Load(iniPath)
	if err != nil {
		logger.Warn("INI not found; Get information from Env variables")
		envName, bootErr := bootstrapFromEnv(iniPath, optionalEnv...)
		if bootErr != nil {
			logger.Error(fmt.Sprintf("Bootstrap failed: %v", bootErr))
			if envName == "" {
				envName = resolveEnvName(optionalEnv...)
			}
			viper.Set(keys.CurrentEnvironment, envName)
			return nil
		}
		cfg, err = ini.Load(iniPath)
		if err != nil {
			logger.Error(fmt.Sprintf("INI written but cannot reload: %v (ENV-only mode)", err))
			viper.Set(keys.CurrentEnvironment, viper.GetString(keys.CurrentEnvironment))
			return nil
		}
	}

	// active env: --env > DEFAULT.current_environment > dhcore_name > default
	env := resolveEnvName(optionalEnv...)
	if env == "default" {
		if v := cfg.Section("DEFAULT").Key("current_environment").String(); v != "" {
			env = v
		}
	}

	if err := loadIniSectionIntoViper(cfg, env); err != nil {
		return fmt.Errorf("failed to load INI into viper: %w", err)
	}
	viper.Set(keys.CurrentEnvironment, env)
	return nil
}

// bootstrapFromEnv creates a new INI by fetching .well-known endpoints when no
// INI file exists yet but DHCORE_ENDPOINT is available in the environment.
// AutomaticEnv (set by SetupViperEnv) makes DHCORE_ENDPOINT available to Viper.
func bootstrapFromEnv(iniPath string, optionalEnv ...string) (string, error) {
	baseEndpoint := viper.GetString(keys.DhCoreEndpoint)
	if baseEndpoint == "" {
		return "", fmt.Errorf("missing %s: set it in env or run 'dhcli register'", keys.DhCoreEndpoint)
	}

	var additionalKeys []string

	cfg, err := FetchConfig(baseEndpoint + "/.well-known/configuration")
	if err != nil {
		return "", fmt.Errorf("fetching configuration failed: %w", err)
	}
	for k, v := range cfg {
		viper.Set(k, ReflectValue(v))
		additionalKeys = append(additionalKeys, k)
	}

	oidc, err := FetchConfig(baseEndpoint + "/.well-known/openid-configuration")
	if err != nil {
		logger.Warn(fmt.Sprintf("OpenID config fetch failed (non-fatal): %v", err))
	} else {
		for k, v := range oidc {
			pk := "oauth2_" + k
			viper.Set(pk, ReflectValue(v))
			additionalKeys = append(additionalKeys, pk)
		}
	}

	envName := resolveEnvName(optionalEnv...)
	if envName == "default" {
		if nm := viper.GetString(keys.DhCoreName); nm != "" {
			envName = nm
		}
	}
	viper.Set(keys.CurrentEnvironment, envName)

	ts := time.Now().UTC().Format(time.RFC3339)
	viper.Set(keys.UpdatedEnvKey, ts)
	additionalKeys = append(additionalKeys, keys.UpdatedEnvKey)

	viper.Set(keys.IniSource, "well-known")
	additionalKeys = append(additionalKeys, keys.IniSource)

	if err := PersistToIni(iniPath, envName, additionalKeys); err != nil {
		return "", fmt.Errorf("write ini failed: %w", err)
	}

	if _, err := ini.Load(iniPath); err != nil {
		return "", fmt.Errorf("ini written but cannot reload: %w", err)
	}

	return envName, nil
}
