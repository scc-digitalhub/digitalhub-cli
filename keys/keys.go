// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
// SPDX-License-Identifier: Apache-2.0

// Package keys defines the canonical Viper/config key names and shared lookup
// maps used across the whole dhcli codebase.
package keys

const (
	IniName            = ".dhcore.ini"
	IniSource          = "ini_source"
	CurrentEnvironment = "current_environment"
	UpdatedEnvKey      = "updated_environment"
	ApiLevelKey        = "dhcore_api_level"

	DhCoreName                  = "dhcore_name"
	DhCoreIssuer                = "dhcore_issuer"
	DhCoreClientId              = "dhcore_client_id"
	DhCoreEndpoint              = "dhcore_endpoint"
	DhCoreApiVersion            = "dhcore_api_version"
	DhCoreAccessToken           = "dhcore_access_token"
	DhCoreIdToken               = "dhcore_id_token"
	DhCoreExpiresAt             = "dhcore_expires_at"
	CredentialsList             = "credentials_list"
	DhCoreUser                  = "dhcore_user"
	DhCorePassword              = "dhcore_password"
	DhCoreRefreshToken          = "dhcore_refresh_token"
	DhCoreProxy                 = "dhcore_proxy"
	OAuth2TokenEndpoint         = "oauth2_token_endpoint"
	OAuth2AuthorizationEndpoint = "oauth2_authorization_endpoint"
	OAuth2ScopesSupported       = "oauth2_scopes_supported"

	// API level the current version of the CLI was developed for
	MinApiLevel = 10

	// API level required for individual commands; 0 means no restriction
	LoginMin   = 10
	LoginMax   = 0
	CreateMin  = 10
	CreateMax  = 0
	ListMin    = 10
	ListMax    = 0
	GetMin     = 10
	GetMax     = 0
	UpdateMin  = 10
	UpdateMax  = 0
	DeleteMin  = 10
	DeleteMax  = 0
	StopMin    = 10
	StopMax    = 0
	LogMin     = 10
	LogMax     = 0
	MetricsMin = 10
	MetricsMax = 0
)

// Resources maps plural resource names to their accepted aliases.
var Resources = map[string][]string{
	"artifacts": {"artifact"},
	"dataitems": {"dataitem"},
	"functions": {"function", "fn"},
	"models":    {"model"},
	"projects":  {"project"},
	"runs":      {"run"},
	"workflows": {"workflow"},
	"logs":      {"log"},
}
