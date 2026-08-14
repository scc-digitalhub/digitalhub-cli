// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"strings"

	"dhcli/handlers/utils"
	"dhcli/keys"

	"github.com/spf13/viper"
	"gopkg.in/ini.v1"
)

// ── Well-known provider filter/transform functions ───────────────────────────

// snakeToCamel converts a snake_case string to CamelCase.
// e.g. "s3_access_key" → "S3AccessKey".
func snakeToCamel(s string) string {
	parts := strings.Split(s, "_")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + strings.ToLower(p[1:])
		}
	}
	return strings.Join(parts, "")
}

// filterS3 keeps entries whose key starts with "s3_" or "aws_".
// When format is "json" keys are renamed to CamelCase (e.g. "s3_access_key" → "S3AccessKey");
// otherwise the original key names are preserved.
func filterS3(entries map[string]string, format string) map[string]string {
	result := make(map[string]string)
	for k, v := range entries {
		lower := strings.ToLower(k)
		if strings.HasPrefix(lower, "s3_") {
			outKey := k
			if format == "json" {
				//convert to CamelCase (e.g. "s3_endpoint" → "S3Endpoint")
				outKey = snakeToCamel(lower)
			}
			result[outKey] = v
		}
		if strings.HasPrefix(lower, "aws_") {
			outKey := k
			if format == "json" {
				// remove the prefix and convert to CamelCase (e.g. "aws_access_key" → "AccessKey")
				lower = strings.TrimPrefix(lower, "aws_")
				if lower == "credentials_expiration" {
					lower = "expiration"
				}
				outKey = snakeToCamel(lower)
			}
			result[outKey] = v
		}
	}

	// for json output, if not empty include version=1 as well
	if format == "json" && len(result) > 0 {
		result["Version"] = "1"
	}

	return result
}

// filterOAuth2 filters oauth2/oidc configuration entries.
func filterOAuth2(entries map[string]string, format string) map[string]string {
	result := make(map[string]string)
	for k, v := range entries {
		lower := strings.ToLower(k)
		if strings.HasPrefix(lower, "oauth2_") || strings.HasPrefix(lower, "oidc_") {
			//remove the prefix
			if strings.HasPrefix(lower, "oauth2_") {
				lower = strings.TrimPrefix(lower, "oauth2_")
			} else if strings.HasPrefix(lower, "oidc_") {
				lower = strings.TrimPrefix(lower, "oidc_")
			}
			result[lower] = v
		}
	}

	return result
}

// filterDB keeps entries whose key starts with "db_" and renames keys to
// ALL_CAPS (e.g. "db_host" → "DB_HOST").
func filterDB(entries map[string]string, format string) map[string]string {
	result := make(map[string]string)
	for k, v := range entries {
		lower := strings.ToLower(k)
		if strings.HasPrefix(lower, "db_") {
			result[strings.ToUpper(lower)] = v
		}
	}
	return result
}

// wellKnownProviders maps provider names to custom filter+transform functions.
// Each function receives the full entries map and the output format string.
// Add new entries here to register additional well-known providers.
var wellKnownProviders = map[string]func(map[string]string, string) map[string]string{
	"s3":     filterS3,
	"db":     filterDB,
	"oauth2": filterOAuth2,
	"oidc":   filterOAuth2,
}

// applyProviderFilter filters and optionally transforms entries for the given
// provider name (already normalised to lowercase) and output format.
// Well-known providers use their registered function; unknown providers do a
// plain prefix filter with no key transformation.
func applyProviderFilter(entries map[string]string, provider, format string) map[string]string {
	if provider == "" {
		return entries
	}
	if fn, ok := wellKnownProviders[provider]; ok {
		return fn(entries, format)
	}
	// Generic fallback: keep keys whose lowercase form starts with "<provider>_".
	pfx := provider + "_"
	result := make(map[string]string)
	for k, v := range entries {
		if strings.HasPrefix(strings.ToLower(k), pfx) {
			result[k] = v
		}
	}
	return result
}

// ── Getters ───────────────────────────────────────────────────────────────────

// getConfigEntriesByProvider returns configuration entries filtered (and
// optionally transformed) by provider and output format. If provider is empty
// all entries are returned unchanged.
func getConfigEntriesByProvider(provider, format string) map[string]string {
	all := make(map[string]string)

	credSet := credentialKeySet()

	cfg, err := ini.Load(utils.GetIniPath())
	if err != nil {
		return all
	}

	env := viper.GetString(keys.CurrentEnvironment)
	if env == "" {
		env = "default"
	}

	for _, k := range cfg.Section(env).Keys() {
		name := k.Name()
		if credSet[name] || utils.InternalKeys[name] || name == keys.CredentialsList {
			continue
		}
		all[name] = viper.GetString(name)
	}

	norm := strings.ToLower(strings.TrimSpace(provider))
	return applyProviderFilter(all, norm, format)
}

// getCredentialEntriesByProvider returns credential entries filtered (and
// optionally transformed) by provider and output format. If provider is empty
// all entries are returned unchanged.
func getCredentialEntriesByProvider(provider, format string) map[string]string {
	all := make(map[string]string)
	for _, k := range utils.SplitCSV(viper.GetString(keys.CredentialsList)) {
		all[k] = viper.GetString(k)
	}
	norm := strings.ToLower(strings.TrimSpace(provider))
	return applyProviderFilter(all, norm, format)
}

// credentialKeySet returns the set of keys listed in credentials_list.
func credentialKeySet() map[string]bool {
	set := make(map[string]bool)
	for _, k := range utils.SplitCSV(viper.GetString(keys.CredentialsList)) {
		set[k] = true
	}
	return set
}
