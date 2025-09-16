// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/spf13/viper"
	"gopkg.in/ini.v1"
)

// EnvDumpPrefix: optional prefix for env lookup (e.g., "DHCORE")
const EnvDumpPrefix = ""

// Config holds all logical keys. Tags:
// - vkey: Viper key
// - env: canonical env name (UPPER_SNAKE). If empty, derived from vkey
// - persist: "true" to write the key into the INI
// - default: optional default to set if key is unset
// - secret: "true" if sensitive (not used here, but handy for logging)
type Config struct {
	AuthorizationEndpoint             string `vkey:"authorization_endpoint"               env:"AUTHORIZATION_ENDPOINT"               persist:"true"`
	AwsAccessKeyID                    string `vkey:"aws_access_key_id"                    env:"AWS_ACCESS_KEY_ID"                    persist:"true"  secret:"true"`
	AwsCredentialsExpiration          string `vkey:"aws_credentials_expiration"           env:"AWS_CREDENTIALS_EXPIRATION"           persist:"true"`
	AwsEndpointURL                    string `vkey:"aws_endpoint_url"                     env:"AWS_ENDPOINT_URL"                     persist:"true"`
	AwsRegion                         string `vkey:"aws_region"                           env:"AWS_REGION"                           persist:"true"`
	AwsSecretAccessKey                string `vkey:"aws_secret_access_key"                env:"AWS_SECRET_ACCESS_KEY"                persist:"true"  secret:"true"`
	AwsSessionToken                   string `vkey:"aws_session_token"                    env:"AWS_SESSION_TOKEN"                    persist:"true"  secret:"true"`
	DbDatabase                        string `vkey:"db_database"                          env:"DB_DATABASE"                          persist:"true"`
	DbHost                            string `vkey:"db_host"                              env:"DB_HOST"                              persist:"true"`
	DbPassword                        string `vkey:"db_password"                          env:"DB_PASSWORD"                          persist:"true"  secret:"true"`
	DbPlatform                        string `vkey:"db_platform"                          env:"DB_PLATFORM"                          persist:"true"`
	DbPort                            string `vkey:"db_port"                              env:"DB_PORT"                              persist:"true"`
	DbUsername                        string `vkey:"db_username"                          env:"DB_USERNAME"                          persist:"true"`
	DhProjects                        string `vkey:"dh_projects"                          env:"DH_PROJECTS"                          persist:"true"`
	DhcoreAccessToken                 string `vkey:"dhcore_access_token"                  env:"DHCORE_ACCESS_TOKEN"                  persist:"true"  secret:"true"`
	DhcoreApiLevel                    string `vkey:"dhcore_api_level"                     env:"DHCORE_API_LEVEL"                     persist:"true"`
	DhcoreApiVersion                  string `vkey:"dhcore_api_version"                   env:"DHCORE_API_VERSION"                   persist:"true"  default:"v1"`
	DhcoreAuthenticationMethods       string `vkey:"dhcore_authentication_methods"        env:"DHCORE_AUTHENTICATION_METHODS"        persist:"true"`
	DhcoreClientId                    string `vkey:"dhcore_client_id"                     env:"DHCORE_CLIENT_ID"                     persist:"true"`
	DhcoreDefaultFilesStore           string `vkey:"dhcore_default_files_store"           env:"DHCORE_DEFAULT_FILES_STORE"           persist:"true"`
	DhcoreEndpoint                    string `vkey:"dhcore_endpoint"                      env:"DHCORE_ENDPOINT"                      persist:"true"`
	DhcoreExpiresIn                   string `vkey:"dhcore_expires_in"                    env:"DHCORE_EXPIRES_IN"                    persist:"true"`
	DhcoreIdToken                     string `vkey:"dhcore_id_token"                      env:"DHCORE_ID_TOKEN"                      persist:"true"  secret:"true"`
	DhcoreIssuer                      string `vkey:"dhcore_issuer"                        env:"DHCORE_ISSUER"                        persist:"true"`
	DhcoreName                        string `vkey:"dhcore_name"                          env:"DHCORE_NAME"                          persist:"true"`
	DhcoreRealm                       string `vkey:"dhcore_realm"                         env:"DHCORE_REALM"                         persist:"true"`
	DhcoreRefreshToken                string `vkey:"dhcore_refresh_token"                 env:"DHCORE_REFRESH_TOKEN"                 persist:"true"  secret:"true"`
	DhcoreVersion                     string `vkey:"dhcore_version"                       env:"DHCORE_VERSION"                       persist:"true"`
	GrantTypesSupported               string `vkey:"grant_types_supported"                env:"GRANT_TYPES_SUPPORTED"                persist:"true"`
	Issuer                            string `vkey:"issuer"                               env:"ISSUER"                               persist:"true"`
	JwksUri                           string `vkey:"jwks_uri"                             env:"JWKS_URI"                             persist:"true"`
	ResponseTypesSupported            string `vkey:"response_types_supported"             env:"RESPONSE_TYPES_SUPPORTED"             persist:"true"`
	S3Bucket                          string `vkey:"s3_bucket"                            env:"S3_BUCKET"                            persist:"true"`
	S3PathStyle                       string `vkey:"s3_path_style"                        env:"S3_PATH_STYLE"                        persist:"true"`
	S3SignatureVersion                string `vkey:"s3_signature_version"                 env:"S3_SIGNATURE_VERSION"                 persist:"true"`
	ScopesSupported                   string `vkey:"scopes_supported"                     env:"SCOPES_SUPPORTED"                     persist:"true"`
	TokenEndpoint                     string `vkey:"token_endpoint"                       env:"TOKEN_ENDPOINT"                       persist:"true"`
	TokenEndpointAuthMethodsSupported string `vkey:"token_endpoint_auth_methods_supported" env:"TOKEN_ENDPOINT_AUTH_METHODS_SUPPORTED" persist:"true"`
	UserinfoEndpoint                  string `vkey:"userinfo_endpoint"                    env:"USERINFO_ENDPOINT"                    persist:"true"`

	// Optional bookkeeping keys; add/remove as you like
	UpdatedEnvironment string `vkey:"updated_environment" env:"UPDATED_ENVIRONMENT" persist:"true" bind:"false"`

	// Expose the active env name in Viper (not persisted here)
	CurrentEnvironment string `vkey:"current_environment" env:"CURRENT_ENVIRONMENT" persist:"false"`
}

// resolveEnvName mirrors your previous selection logic.
func resolveEnvName(optionalEnv ...string) string {
	if len(optionalEnv) > 0 && optionalEnv[0] != "" && strings.ToLower(optionalEnv[0]) != "null" {
		return optionalEnv[0]
	}
	if v := os.Getenv("CURRENT_ENVIRONMENT"); v != "" {
		return v
	}
	if v := os.Getenv("DHCORE_NAME"); v != "" {
		return v
	}
	return "env"
}

// mirror PREFIX_FOO -> FOO, so we can BindEnv with one canonical name.
func mirrorPrefix(prefix string) {
	if prefix == "" {
		return
	}
	upPrefix := strings.ToUpper(prefix) + "_"
	for _, e := range os.Environ() {
		kv := strings.SplitN(e, "=", 2)
		if len(kv) != 2 {
			continue
		}
		name, val := kv[0], kv[1]
		if strings.HasPrefix(name, upPrefix) {
			unpref := strings.TrimPrefix(name, upPrefix)
			if os.Getenv(unpref) == "" {
				_ = os.Setenv(unpref, val)
			}
		}
	}
}

// Bind env for all fields of Config using struct tags.
// - Supports prefix via mirrorPrefix()
// - Sets defaults from `default` tag
func BindEnvFromStruct(prefix string) {
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	mirrorPrefix(prefix)

	rt := reflect.TypeOf(Config{})
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)

		key := f.Tag.Get("vkey")
		if key == "" {
			continue
		}

		if f.Tag.Get("bind") == "false" {
			if def := f.Tag.Get("default"); def != "" && !viper.IsSet(key) {
				viper.SetDefault(key, def)
			}
			continue
		}

		env := f.Tag.Get("env")
		if env == "" {
			env = strings.ToUpper(strings.ReplaceAll(key, ".", "_"))
		}
		_ = viper.BindEnv(key, env)

		if def := f.Tag.Get("default"); def != "" && !viper.IsSet(key) {
			viper.SetDefault(key, def)
		}
	}
}

// Write a new INI with only fields marked persist:"true".
func WriteIniFromStruct(iniPath, envName string) error {
	cfg := ini.Empty()
	cfg.Section("DEFAULT").Key("current_environment").SetValue(envName)
	sec := cfg.Section(envName)

	rt := reflect.TypeOf(Config{})
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		if f.Tag.Get("persist") != "true" {
			continue
		}
		key := f.Tag.Get("vkey")
		if key == "" {
			continue
		}
		val := viper.GetString(key)
		if val == "" {
			continue
		}
		sec.Key(key).SetValue(val)
	}

	return cfg.SaveTo(iniPath)
}

// Update or create INI section from current Viper values (persist:"true" only).
func UpdateIniFromStruct(iniPath, envName string) error {
	cfg, err := ini.Load(iniPath)
	if err != nil {
		return WriteIniFromStruct(iniPath, envName)
	}
	sec := cfg.Section(envName)

	rt := reflect.TypeOf(Config{})
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		if f.Tag.Get("persist") != "true" {
			continue
		}
		key := f.Tag.Get("vkey")
		if key == "" {
			continue
		}
		val := viper.GetString(key)
		if val == "" {
			continue
		}
		sec.Key(key).SetValue(val)
	}

	if !cfg.Section("DEFAULT").HasKey("current_environment") {
		cfg.Section("DEFAULT").Key("current_environment").SetValue(envName)
	}
	sec.Key(UpdatedEnvKey).SetValue(time.Now().UTC().Format(time.RFC3339))
	return cfg.SaveTo(iniPath)
}

// Load [DEFAULT] + [env] into Viper (TOML in-memory). ENV can still override on Get().
func loadIniSectionIntoViper(cfg *ini.File, env string) error {
	def := cfg.Section("DEFAULT")
	selected := def
	if env != "" && cfg.HasSection(env) {
		selected = cfg.Section(env)
		fmt.Printf("Using section: [%s]\n", env)
	} else if env == "" || strings.EqualFold(env, "DEFAULT") {
		fmt.Println("no environment selected, using [DEFAULT]")
	} else {
		fmt.Println("current_environment not found/invalid, falling back to [DEFAULT]")
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
// 1) bind ENV from struct (live, no mass Set)
// 2) load INI or bootstrap it (persist only fields persist:"true")
// 3) load active section into Viper and set current_environment
func RegisterIniCfgWithViper(optionalEnv ...string) error {
	iniPath := getIniPath() // assume defined elsewhere in utils

	BindEnvFromStruct(EnvDumpPrefix)

	cfg, err := ini.Load(iniPath)
	if err != nil {
		fmt.Println("INI file not found; bootstrapping from environment variables…")
		envName := resolveEnvName(optionalEnv...)
		if err := WriteIniFromStruct(iniPath, envName); err != nil {
			fmt.Printf("failed to create ini from env (%v); continuing in env-only mode\n", err)
			viper.Set("current_environment", envName)
			return nil
		}
		cfg, err = ini.Load(iniPath)
		if err != nil {
			fmt.Printf("created ini but cannot read it back (%v); continuing in env-only mode\n", err)
			viper.Set("current_environment", envName)
			return nil
		}
	}

	env := ""
	if len(optionalEnv) > 0 && optionalEnv[0] != "" && strings.ToLower(optionalEnv[0]) != "null" {
		env = optionalEnv[0]
	} else {
		env = cfg.Section("DEFAULT").Key("current_environment").String()
		if env == "" {
			env = resolveEnvName() // fallback if DEFAULT lacks it
		}
	}

	if err := loadIniSectionIntoViper(cfg, env); err != nil {
		return fmt.Errorf("failed to load INI into viper: %w", err)
	}
	viper.Set("current_environment", env)
	return nil
}
