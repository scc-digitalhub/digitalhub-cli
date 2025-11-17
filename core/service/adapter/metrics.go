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

func MetricsHandler(env string, project string, container string, resource string, id string) error {
	endpoint := utils.TranslateEndpoint(resource)

	// Load environment and check API level requirements
	utils.CheckUpdateEnvironment()
	utils.CheckApiLevel(utils.ApiLevelKey, utils.MetricsMin, utils.MetricsMax)

	if project == "" {
		return errors.New("Project not specified.")
	}

	cfg := sdk.Config{
		Core: sdk.CoreConfig{
			BaseURL:     viper.GetString(utils.DhCoreEndpoint),
			APIVersion:  viper.GetString(utils.DhCoreApiVersion),
			AccessToken: viper.GetString(utils.DhCoreAccessToken),
		},
	}

	svc, err := sdk.NewMetricsService(context.Background(), cfg)
	if err != nil {
		return err
	}

	req := sdk.MetricsRequest{
		Project:   project,
		Endpoint:  endpoint,
		ID:        id,
		Container: container,
	}

	// Same semantics as old MetricsHandler: just print or say "No metrics..."
	if err := svc.PrintMetrics(context.Background(), req); err != nil {
		return err
	}

	return nil
}
