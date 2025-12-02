// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"context"
	"dhcli/sdk/config"
	runsvc "dhcli/sdk/services/run"
	"dhcli/sdk/utils"
	"errors"

	"github.com/spf13/viper"
)

func MetricsHandler(env string, project string, container string, resource string, id string) error {
	endpoint := utils.TranslateEndpoint(resource)

	// Checks identici all’originale
	utils.CheckUpdateEnvironment()
	utils.CheckApiLevel(utils.ApiLevelKey, utils.MetricsMin, utils.MetricsMax)

	if project == "" {
		return errors.New("project not specified")
	}

	// Adapter → sdk.Config
	cfg := config.Config{
		Core: config.CoreConfig{
			BaseURL:     viper.GetString(utils.DhCoreEndpoint),
			APIVersion:  viper.GetString(utils.DhCoreApiVersion),
			AccessToken: viper.GetString(utils.DhCoreAccessToken),
		},
	}

	ctx := context.Background()

	// RunService globale
	svc, err := runsvc.NewRunService(ctx, cfg)
	if err != nil {
		return err
	}

	// Nuova request coerente col RunService
	req := runsvc.MetricsRequest{
		RunResourceRequest: runsvc.RunResourceRequest{
			Project:  project,
			Endpoint: endpoint,
			ID:       id,
		},
		Container: container,
	}

	// Comportamento identico all’originale:
	// stampa metrics o “No metrics for this run.”
	if err := svc.PrintMetrics(ctx, req); err != nil {
		return err
	}

	return nil
}
