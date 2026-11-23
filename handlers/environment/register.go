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
	endpoint = strings.TrimRight(endpoint, "/")
	if endpoint == "" {
		return fmt.Errorf("endpoint is required")
	}
	endpoint += "/"

	cfg, err := utils.LoadIni(true)
	if err != nil {
		return err
	}

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
	sameEndpoint := false
	if cfg.HasSection(env) {
		existingSection := cfg.Section(env)
		if existingSection.HasKey(keys.DhCoreEndpoint) {
			existingEndpoint := existingSection.Key(keys.DhCoreEndpoint).String()
			// Normalize both endpoints by removing trailing slashes for comparison
			normalizedExisting := strings.TrimRight(existingEndpoint, "/")
			normalizedNew := strings.TrimRight(endpoint, "/")
			sameEndpoint = normalizedExisting != "" && normalizedExisting == normalizedNew
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

	// Preserve locally stored credentials and settings only for the same endpoint.
	section := cfg.Section(env)
	credentials := map[string]string{}
	if sameEndpoint {
		var dynamicCreds []string
		if key, err := section.GetKey(keys.CredentialsList); err == nil {
			dynamicCreds = utils.SplitCSV(key.String())
		}
		credentialKeys := append(dynamicCreds,
			keys.CredentialsList, keys.DhCoreAccessToken, keys.DhCoreRefreshToken,
			keys.DhCoreIdToken, keys.DhCoreExpiresAt,
			"aws_access_key_id", "aws_secret_access_key", "aws_session_token")
		for _, name := range credentialKeys {
			if section.HasKey(name) {
				credentials[name] = section.Key(name).String()
			}
		}
	}
	for _, k := range section.Keys() {
		section.DeleteKey(k.Name())
	}

	// 4. Reflect config keys
	for k, v := range config {
		section.Key(k).SetValue(utils.ReflectValue(v))
	}
	section.Key(keys.DhCoreEndpoint).SetValue(strings.TrimRight(endpoint, "/"))

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

		// All OIDC keys are prefixed with oauth2_ to namespace them.
		targetKey := "oauth2_" + k
		valStr := utils.ReflectValue(v)
		section.Key(targetKey).SetValue(valStr)
	}

	// 7. Add timestamp
	section.Key(keys.UpdatedEnvKey).SetValue(time.Now().UTC().Format(time.RFC3339))

	// 8. Add ini_source
	section.Key(keys.IniSource).SetValue("well-known")
	for name, value := range credentials {
		section.Key(name).SetValue(value)
	}

	// 9. Set default as default environment if not already set
	defaultSection := cfg.Section("DEFAULT")
	if defaultSection.HasKey(keys.CurrentEnvironment) {
		defaultSection.DeleteKey(keys.CurrentEnvironment)
	}
	defaultSection.NewKey(keys.CurrentEnvironment, env)

	if err := utils.SaveIni(cfg); err != nil {
		return err
	}

	log.Printf("'%v' registered.\n", env)
	return nil
}
