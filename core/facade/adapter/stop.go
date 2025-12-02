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

func StopHandler(env string, project string, resource string, id string) error {
	endpoint := utils.TranslateEndpoint(resource)

	// Preserve original guards/compat behavior
	utils.CheckUpdateEnvironment()
	utils.CheckApiLevel(utils.ApiLevelKey, utils.StopMin, utils.StopMax)

	if project == "" {
		return errors.New("project not specified")
	}

	// Adapter: viper → sdk.Config
	cfg := config.Config{
		Core: config.CoreConfig{
			BaseURL:     viper.GetString(utils.DhCoreEndpoint),
			APIVersion:  viper.GetString(utils.DhCoreApiVersion),
			AccessToken: viper.GetString(utils.DhCoreAccessToken),
		},
	}

	ctx := context.Background()

	// Usa il nuovo RunService globale al posto del vecchio StopService
	svc, err := runsvc.NewRunService(ctx, cfg)
	if err != nil {
		return err
	}

	// Request adattata al nuovo sistema (RunResourceRequest embedded)
	respBody, _, err := svc.Stop(ctx, runsvc.StopRequest{
		RunResourceRequest: runsvc.RunResourceRequest{
			Project:  project,
			Endpoint: endpoint,
			ID:       id,
		},
	})
	if err != nil {
		return err
	}

	// Mantieniamo comportamento originale: stampa lo stato
	return utils.PrintResponseState(respBody)
}
