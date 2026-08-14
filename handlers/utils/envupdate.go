// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"fmt"
	"time"

	"dhcli/keys"

	"github.com/spf13/viper"
)

// CheckUpdateEnvironment decides whether to refresh the environment:
// - endpoint present but api_level absent -> bootstrap (well-known never fetched)
// - missing/empty timestamp               -> update
// - invalid timestamp                     -> update
// - older than TTL                        -> update
// - otherwise                             -> skip (fresh)
func CheckUpdateEnvironment() {
	const key = keys.UpdatedEnvKey

	endpoint := viper.GetString(keys.DhCoreEndpoint)
	apiLevel := viper.GetString(keys.ApiLevelKey)

	// Partial config: endpoint known but well-known endpoints never fetched.
	if endpoint != "" && apiLevel == "" {
		logger.Warn("Config has endpoint but no api_level — bootstrapping from well-known.")
		updateEnvironment()
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
	baseEndpoint := viper.GetString(keys.DhCoreEndpoint)
	if baseEndpoint == "" {
		logger.Warn("Skip: dhcore_endpoint is empty.")
		return
	}

	var additionalKeys []string

	cfg, err := FetchConfig(baseEndpoint + "/.well-known/configuration")
	if err != nil {
		logger.Error(fmt.Sprintf("Config fetch failed: %v", err))
		return
	}
	for k, v := range cfg {
		viper.Set(k, ReflectValue(v))
		additionalKeys = append(additionalKeys, k)
	}

	oidc, err := FetchConfig(baseEndpoint + "/.well-known/openid-configuration")
	if err != nil {
		logger.Error(fmt.Sprintf("OpenID fetch failed: %v", err))
		return
	}
	for k, v := range oidc {
		pk := "oauth2_" + k
		viper.Set(pk, ReflectValue(v))
		additionalKeys = append(additionalKeys, pk)
	}

	ts := time.Now().UTC().Format(time.RFC3339)
	viper.Set(keys.UpdatedEnvKey, ts)
	additionalKeys = append(additionalKeys, keys.UpdatedEnvKey)
	logger.Info(fmt.Sprintf("Set %s=%s", keys.UpdatedEnvKey, ts))

	env := viper.GetString(keys.CurrentEnvironment)
	if env == "" {
		env = resolveEnvName()
	}
	if err := PersistToIni(GetIniPath(), env, additionalKeys); err != nil {
		logger.Warn(fmt.Sprintf("Persist skipped (read-only or missing ini): %v", err))
	} else {
		logger.Info(fmt.Sprintf("Persisted to [%s].", env))
	}
}

// UpdateIniSectionFromViper persists the current environment section to the INI.
// The additionalKeys argument is accepted for backward-compatibility but ignored.
func UpdateIniSectionFromViper(_ []string) error {
	return PersistCurrentEnv(nil)
}
