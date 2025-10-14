// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package sdk

// Config complessiva passata all’SDK (niente viper/INI qui)
type Config struct {
	Core CoreConfig
	S3   S3Config
}

type CoreConfig struct {
	BaseURL     string
	APIVersion  string
	AccessToken string
}

type S3Config struct {
	AccessKey    string
	SecretKey    string
	SessionToken string
	Region       string
	EndpointURL  string
}
