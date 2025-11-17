// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"context"
	"dhcli/sdk"
	"dhcli/utils"
	"errors"

	"github.com/spf13/viper"
)

func ResumeHandler(env string, project string, resource string, id string) error {
	endpoint := utils.TranslateEndpoint(resource)

	// Preserve original guards / API level checks
	utils.CheckUpdateEnvironment()
	utils.CheckApiLevel(utils.ApiLevelKey, utils.ResumeMin, utils.ResumeMax)

	if project == "" {
		return errors.New("project not specified")
	}

	cfg := sdk.Config{
		Core: sdk.CoreConfig{
			BaseURL:     viper.GetString(utils.DhCoreEndpoint),
			APIVersion:  viper.GetString(utils.DhCoreApiVersion),
			AccessToken: viper.GetString(utils.DhCoreAccessToken),
		},
	}

	svc, err := sdk.NewResumeService(context.Background(), cfg)
	if err != nil {
		return err
	}

	respBody, _, err := svc.Resume(context.Background(), sdk.ResumeRequest{
		Project:  project,
		Endpoint: endpoint,
		ID:       id,
	})
	if err != nil {
		return err
	}

	return utils.PrintResponseState(respBody)
}
