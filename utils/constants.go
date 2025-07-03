// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package utils

const (
	IniName            = ".dhcore.ini"
	CurrentEnvironment = "current_environment"
	configFile         = "config.json"
	ApiLevelKey        = "dhcore_api_level"
	ClientIdKey        = "dhcore_client_id"
	UpdatedEnvKey      = "updated_environment"
	DhCoreEndpoint     = "dhcore_endpoint"

	outdatedAfterHours = 1

	// API level the current version of the CLI was developed for
	MinApiLevel = 10

	// API level required for individual commands; 0 means no restriction
	LoginMin  = 10
	LoginMax  = 0
	CreateMin = 10
	CreateMax = 0
	ListMin   = 10
	ListMax   = 0
	GetMin    = 10
	GetMax    = 0
	UpdateMin = 10
	UpdateMax = 0
	DeleteMin = 10
	DeleteMax = 0
	StopMin   = 10
	StopMax   = 0
	ResumeMin = 10
	ResumeMax = 0
	LogMin    = 10
	LogMax    = 0
)

var OpenIdFields = []string{"authorization_endpoint", "token_endpoint", "issuer", "scopes_supported", "access_token", "refresh_token"}
