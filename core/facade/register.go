// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package facade

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/scc-digitalhub/digitalhub-cli-sdk/sdk/utils"
)

func RegisterHandler(env string, endpoint string) error {
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
		env = utils.GetStringValue(config, utils.DhCoreName)
		if env == "" {
			return fmt.Errorf("environment not specified and not defined in core configuration")
		}
	}

	// 2. Clear section if it exists
	if cfg.HasSection(env) {
		log.Printf("Section '%v' already exists, will be overwritten.\n", env)
	}
	section := cfg.Section(env)
	for _, k := range section.Keys() {
		section.DeleteKey(k.Name())
	}

	// 3. Reflect config keys
	for k, v := range config {
		section.NewKey(k, utils.ReflectValue(v))
	}

	// 4. Check API level
	apiLevel := utils.GetStringValue(config, utils.ApiLevelKey)
	apiLevelInt, err := strconv.Atoi(apiLevel)
	if err != nil {
		log.Println("WARNING: API level not valid or missing.")
	} else if apiLevelInt < utils.MinApiLevel {
		log.Printf("WARNING: API level %v < minimum required %v\n", apiLevelInt, utils.MinApiLevel)
	}

	// 5. Fetch and reflect OpenID config
	openIdConfig, err := utils.FetchConfig(endpoint + ".well-known/openid-configuration")
	if err != nil {
		return fmt.Errorf("fetching OpenID configuration failed: %w", err)
	}

	keys := make([]string, 0, len(openIdConfig))
	for k := range openIdConfig {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := openIdConfig[k]

		// remap only if there is a DHCORE correspondence
		targetKey := k
		if dhKey, has := utils.DhCoreMap[k]; has {
			targetKey = dhKey
		}

		// ReflectValue deve gestire slice/array (es. scopes_supported)
		valStr := utils.ReflectValue(v)

		// section.Key crea se non esiste; SetValue sovrascrive/assegna
		section.Key(targetKey).SetValue(valStr)
	}

	// 6. Add timestamp
	section.NewKey(utils.UpdatedEnvKey, time.Now().UTC().Format(time.RFC3339))

	// 7. Add ini_source
	section.NewKey(utils.IniSource, "well-known")

	// 8. Set default env if missing
	defaultSection := cfg.Section("DEFAULT")
	if !defaultSection.HasKey(utils.CurrentEnvironment) {
		defaultSection.NewKey(utils.CurrentEnvironment, env)
	}

	utils.SaveIni(cfg)

	log.Printf("'%v' registered.\n", env)
	return nil
}
