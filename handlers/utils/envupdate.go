// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// CheckUpdateEnvironment decides whether to refresh the environment:
// - missing/empty timestamp -> update
// - invalid timestamp       -> update
// - older than TTL          -> update
func CheckUpdateEnvironment() {
	const key = UpdatedEnvKey

	if viper.IsSet(IniSource) && viper.GetString(IniSource) == "env" {
		logger.Info("INI file has been created from enviromental variables...skip update")

		return
	}

	val := viper.GetString(key)
	isSet := viper.IsSet(key)
	logger.Step(fmt.Sprintf("Config freshness (%s): isSet=%v value=%q", key, isSet, val))

	if !isSet || val == "" {
		logger.Warn("Update: no timestamp.")
		updateEnvironment()
		return
	}

	t, err := time.Parse(time.RFC3339, val)
	if err != nil {
		logger.Warn(fmt.Sprintf("Update: invalid timestamp (%v).", err))
		updateEnvironment()
		return
	}

	now := time.Now().UTC()
	age := now.Sub(t.UTC())
	ttl := time.Duration(outdatedAfterHours) * time.Hour

	if age >= ttl {
		logger.Step(fmt.Sprintf("Update: outdated (age %s ≥ TTL %s).", age, ttl))
		updateEnvironment()
		return
	}

	logger.Step(fmt.Sprintf("Fresh: age %s < TTL %s.", age, ttl))
}

// Fetch well-known, update Viper, bump timestamp, persist allowlisted keys.
func updateEnvironment() {
	logger.Info("Updating environment…")
	baseEndpoint := viper.GetString(DhCoreEndpoint)
	if baseEndpoint == "" {
		logger.Warn("Skip: dhcore_endpoint is empty.")
		return
	}

	cfg, err := FetchConfig(baseEndpoint + "/.well-known/configuration")
	if err != nil {
		logger.Error(fmt.Sprintf("Config fetch failed: %v", err))
		return
	}
	for k, v := range cfg {
		viper.Set(k, ReflectValue(v))
	}

	oidc, err := FetchConfig(baseEndpoint + "/.well-known/openid-configuration")
	if err != nil {
		logger.Error(fmt.Sprintf("OpenID fetch failed: %v", err))
		return
	}
	for k, v := range oidc {
		viper.Set(k, ReflectValue(v))
	}

	ts := time.Now().UTC().Format(time.RFC3339)
	viper.Set(UpdatedEnvKey, ts)
	logger.Info(fmt.Sprintf("Set %s=%s", UpdatedEnvKey, ts))

	env := viper.GetString(CurrentEnvironment)
	if env == "" {
		env = resolveEnvName()
	}
	if err := UpdateIniFromStruct(getIniPath(), env); err != nil {
		logger.Error(fmt.Sprintf("Persist failed: %v", err))
		return
	}
	logger.Info(fmt.Sprintf("Persisted to [%s].", env))
}

// Backward-compat wrapper.
func UpdateIniSectionFromViper(_ []string) error {
	env := viper.GetString(CurrentEnvironment)
	if env == "" {
		env = resolveEnvName()
	}
	if err := UpdateIniFromStruct(getIniPath(), env); err != nil {
		return fmt.Errorf("failed to save ini: %w", err)
	}
	logger.Info(fmt.Sprintf("Updated section [%s] in %s", env, getIniPath()))
	return nil
}
