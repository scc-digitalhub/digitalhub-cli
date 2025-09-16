// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package utils

const (
	IniName            = ".dhcore.ini"
	CurrentEnvironment = "current_environment"
	UpdatedEnvKey      = "updated_environment"
	ApiLevelKey        = "dhcore_api_level"
	DhCoreIssuer       = "dhcore_issuer"
	DhCoreClientId     = "dhcore_client_id"
	DhCoreEndpoint     = "dhcore_endpoint"
	DhCoreApiVersion   = "dhcore_api_version"
	DhCoreAccessToken  = "dhcore_access_token"
	DhCoreRefreshToken = "dhcore_refresh_token"

	outdatedAfterHours = 1

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
	ResumeMin  = 10
	ResumeMax  = 0
	LogMin     = 10
	LogMax     = 0
	MetricsMin = 10
	MetricsMax = 0
)

var DhCoreMap = map[string]string{
	"issuer":             DhCoreIssuer,
	"client_id":          DhCoreClientId,
	"dhcore_endpoint":    DhCoreEndpoint,
	"dhcore_api_version": DhCoreApiVersion,
	"access_token":       DhCoreAccessToken,
	"refresh_token":      DhCoreRefreshToken,
}

var Resources = map[string][]string{
	"artifacts": []string{"artifact"},
	"dataitems": []string{"dataitem"},
	"functions": []string{"function", "fn"},
	"models":    []string{"model"},
	"projects":  []string{"project"},
	"runs":      []string{"run"},
	"workflows": []string{"workflow"},
	"logs":      []string{"log"},
}
