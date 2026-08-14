// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"context"
	"errors"

	runsvc "github.com/scc-digitalhub/digitalhub-cli-sdk/sdk/services/run"

	"github.com/scc-digitalhub/digitalhub-cli-sdk/sdk/config"

	"dhcli/handlers/utils"
	"dhcli/keys"

	"github.com/spf13/viper"
)

func StopHandler(env string, project string, id string) error {
	endpoint := "runs"

	// Preserve original guards/compat behavior
	utils.CheckUpdateEnvironment()
	utils.CheckApiLevel(keys.ApiLevelKey, keys.StopMin, keys.StopMax)
	if err := utils.CheckCredentials(); err != nil {
		return err
	}

	if project == "" {
		return errors.New("project not specified")
	}

	// Adapter: viper → sdk.Config
	cfg := config.Config{
		Core: config.CoreConfig{
			BaseURL:     viper.GetString(keys.DhCoreEndpoint),
			APIVersion:  viper.GetString(keys.DhCoreApiVersion),
			AccessToken: viper.GetString(keys.DhCoreAccessToken),
		},
		HTTPClient: utils.GetDebugHTTPClient(),
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
			Resource: endpoint,
			ID:       id,
		},
	})
	if err != nil {
		return err
	}

	// Mantieniamo comportamento originale: stampa lo stato
	return utils.PrintResponseState(respBody)
}
