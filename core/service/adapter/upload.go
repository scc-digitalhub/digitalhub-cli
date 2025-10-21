// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"context"
	"dhcli/sdk"
	"dhcli/utils"
	"errors"
	"fmt"

	"github.com/spf13/viper"
)

func UploadHandler(env string, input string, project string, resource string, id string, name string, verbose bool) error {
	if input == "" {
		return errors.New("missing required input file or directory")
	}

	endpoint := utils.TranslateEndpoint(resource)
	if endpoint != "projects" && project == "" {
		return errors.New("project is mandatory for non-project resources")
	}

	// Traduzione viper -> sdk.Config (retro-compat)
	cfg := sdk.Config{
		Core: sdk.CoreConfig{
			BaseURL:     viper.GetString(utils.DhCoreEndpoint),
			APIVersion:  viper.GetString(utils.DhCoreApiVersion),
			AccessToken: viper.GetString(utils.DhCoreAccessToken),
		},
		S3: sdk.S3Config{
			AccessKey:    viper.GetString("aws_access_key_id"),
			SecretKey:    viper.GetString("aws_secret_access_key"),
			SessionToken: viper.GetString("aws_session_token"),
			Region:       viper.GetString("aws_region"),
			EndpointURL:  viper.GetString("aws_endpoint_url"),
		},
	}

	// bucket override da viper, "datalake" di default.
	bucket := viper.GetString("s3_bucket")
	if bucket == "" {
		bucket = "datalake"
	}

	svc, err := sdk.NewUploadService(context.Background(), cfg)
	if err != nil {
		return fmt.Errorf("sdk init failed: %w", err)
	}

	_, err = svc.Upload(context.Background(), endpoint, sdk.UploadRequest{
		Project:  project,
		Resource: resource,
		ID:       id,
		Name:     name,
		Input:    input,
		Verbose:  verbose,
		Bucket:   bucket,
	})
	if err != nil {
		return err
	}

	return nil
}
