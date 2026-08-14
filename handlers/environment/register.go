// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package environment

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"dhcli/handlers/utils"
	"dhcli/keys"
)

func RegisterHandler(env string, endpoint string, force bool) error {
	if endpoint == "" {
		return fmt.Errorf("endpoint is required")
	}
	if !strings.HasSuffix(endpoint, "/") {
		endpoint += "/"
	}

	cfg := utils.LoadIni(true)

	// 1. Fetch core config
	config, err := utils.FetchConfig(endpoint + ".well-known/configuration")
	if err != nil {
		return fmt.Errorf("fetching configuration failed: %w", err)
	}

	if env == "" || env == "null" {
		env = utils.GetStringValue(config, keys.DhCoreName)
		if env == "" {
			return fmt.Errorf("environment not specified and not defined in core configuration")
		}
	}

	// 2. Check for endpoint conflict before clearing section
	if cfg.HasSection(env) {
		existingSection := cfg.Section(env)
		if existingSection.HasKey(keys.DhCoreEndpoint) {
			existingEndpoint := existingSection.Key(keys.DhCoreEndpoint).String()
			// Normalize both endpoints by removing trailing slashes for comparison
			normalizedExisting := strings.TrimSuffix(existingEndpoint, "/")
			normalizedNew := strings.TrimSuffix(endpoint, "/")
			if normalizedExisting != normalizedNew && !force {
				return fmt.Errorf("environment '%v' already exists with different endpoint: %v (new: %v). Use --force to override", env, existingEndpoint, endpoint)
			}
		}
		if force {
			log.Printf("Section '%v' already exists, overwriting (--force).\n", env)
		} else {
			log.Printf("Section '%v' already exists, will be overwritten.\n", env)
		}
	}

	// 3. Clear section if it exists
	section := cfg.Section(env)
	for _, k := range section.Keys() {
		section.DeleteKey(k.Name())
	}

	// 4. Reflect config keys
	for k, v := range config {
		section.NewKey(k, utils.ReflectValue(v))
	}

	// 5. Check API level
	apiLevel := utils.GetStringValue(config, keys.ApiLevelKey)
	apiLevelInt, err := strconv.Atoi(apiLevel)
	if err != nil {
		log.Println("WARNING: API level not valid or missing.")
	} else if apiLevelInt < keys.MinApiLevel {
		log.Printf("WARNING: API level %v < minimum required %v\n", apiLevelInt, keys.MinApiLevel)
	}

	// 6. Fetch and reflect OpenID config
	openIdConfig, err := utils.FetchConfig(endpoint + ".well-known/openid-configuration")
	if err != nil {
		return fmt.Errorf("fetching OpenID configuration failed: %w", err)
	}

	sortedKeys := make([]string, 0, len(openIdConfig))
	for k := range openIdConfig {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	for _, k := range sortedKeys {
		v := openIdConfig[k]

		// remap only if there is a DHCORE correspondence
		targetKey := k
		if dhKey, has := keys.DhCoreMap[k]; has {
			targetKey = dhKey
		}

		// ReflectValue deve gestire slice/array (es. scopes_supported)
		valStr := utils.ReflectValue(v)

		// section.Key crea se non esiste; SetValue sovrascrive/assegna
		section.Key(targetKey).SetValue(valStr)
	}

	// 7. Add timestamp
	section.NewKey(keys.UpdatedEnvKey, time.Now().UTC().Format(time.RFC3339))

	// 8. Add ini_source
	section.NewKey(keys.IniSource, "well-known")

	// 9. Set default env if missing
	defaultSection := cfg.Section("DEFAULT")
	if !defaultSection.HasKey(keys.CurrentEnvironment) {
		defaultSection.NewKey(keys.CurrentEnvironment, env)
	}

	utils.SaveIni(cfg)

	log.Printf("'%v' registered.\n", env)
	return nil
}
