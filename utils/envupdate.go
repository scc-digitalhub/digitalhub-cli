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
// - invalid timestamp -> update
// - older than TTL -> update
func CheckUpdateEnvironment() {
	const key = UpdatedEnvKey // "updated_environment"

	val := viper.GetString(key)
	isSet := viper.IsSet(key)
	fmt.Printf("⏱️  Checking config freshness (%s)… isSet=%v, value=%q\n", key, isSet, val)

	// 1) Missing or empty
	if !isSet || val == "" {
		fmt.Println("🔄 Update needed: no timestamp found.")
		updateEnvironment()
		return
	}

	// 2) Invalid RFC3339
	t, err := time.Parse(time.RFC3339, val)
	if err != nil {
		fmt.Printf("🔄 Update needed: invalid timestamp (%v).\n", err)
		updateEnvironment()
		return
	}

	// 3) Outdated
	now := time.Now().UTC()
	ut := t.UTC()
	age := now.Sub(ut)
	ttl := time.Duration(outdatedAfterHours) * time.Hour

	if age >= ttl {
		fmt.Printf("🔄 Update needed: outdated (age %s ≥ TTL %s).\n", age, ttl)
		updateEnvironment()
		return
	}

	fmt.Printf("✅ Fresh: age %s < TTL %s. No update.\n", age, ttl)
}

// updateEnvironment fetches well-known configs, updates Viper, bumps the timestamp,
// and persists only allowlisted keys (struct+tag) into the INI.
func updateEnvironment() {
	fmt.Println("🔁 Updating environment…")
	baseEndpoint := viper.GetString(DhCoreEndpoint)
	if baseEndpoint == "" {
		// Probabilmente RegisterIniCfgWithViper non ha ancora caricato l'endpoint.
		fmt.Println("⚠️  Skip: dhcore_endpoint is empty.")
		return
	}

	// 1) Core configuration
	config, err := FetchConfig(baseEndpoint + "/.well-known/configuration")
	if err != nil {
		fmt.Printf("✗ Config fetch failed: %v\n", err)
		return
	}
	for k, v := range config {
		viper.Set(k, ReflectValue(v))
	}

	// 2) OpenID configuration
	openIdConfig, err := FetchConfig(baseEndpoint + "/.well-known/openid-configuration")
	if err != nil {
		fmt.Printf("✗ OpenID config fetch failed: %v\n", err)
		return
	}
	for k, v := range openIdConfig {
		viper.Set(k, ReflectValue(v))
	}

	// 3) Timestamp (UTC, RFC3339)
	ts := time.Now().UTC().Format(time.RFC3339)
	viper.Set(UpdatedEnvKey, ts)
	fmt.Printf("🕒 Set %s=%s (UTC)\n", UpdatedEnvKey, ts)

	// 4) Persist ONLY allowlisted keys into the current section
	env := viper.GetString(CurrentEnvironment)
	if env == "" {
		env = resolveEnvName() // fallback prudente
	}
	if err := UpdateIniFromStruct(getIniPath(), env); err != nil {
		fmt.Printf("⚠️  Persist failed: %v\n", err)
		return
	}
	fmt.Printf("💾 Persisted allowlisted to section [%s].\n", env)
}

// UpdateIniSectionFromViper is kept for backward compatibility.
// It now delegates to UpdateIniFromStruct, ignoring the provided keys.
func UpdateIniSectionFromViper(_ []string) error {
	env := viper.GetString(CurrentEnvironment)
	if env == "" {
		env = resolveEnvName()
	}
	if err := UpdateIniFromStruct(getIniPath(), env); err != nil {
		return fmt.Errorf("failed to save ini (allowlisted): %w", err)
	}
	fmt.Printf("💾 Updated section [%s] in %s\n", env, getIniPath())
	return nil
}
