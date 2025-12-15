// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"context"
	"errors"

	runsvc "github.com/scc-digitalhub/digitalhub-cli-sdk/sdk/services/run"

	"github.com/scc-digitalhub/digitalhub-cli-sdk/sdk/config"

	"github.com/scc-digitalhub/digitalhub-cli-sdk/sdk/utils"

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

	cfg := config.Config{
		Core: config.CoreConfig{
			BaseURL:     viper.GetString(utils.DhCoreEndpoint),
			APIVersion:  viper.GetString(utils.DhCoreApiVersion),
			AccessToken: viper.GetString(utils.DhCoreAccessToken),
		},
	}

	ctx := context.Background()

	// RunService globale al posto del vecchio ResumeService
	svc, err := runsvc.NewRunService(ctx, cfg)
	if err != nil {
		return err
	}

	// Usa la nuova ResumeRequest con la RunResourceRequest embedded
	respBody, _, err := svc.Resume(ctx, runsvc.ResumeRequest{
		RunResourceRequest: runsvc.RunResourceRequest{
			Project:  project,
			Endpoint: endpoint,
			ID:       id,
		},
	})
	if err != nil {
		return err
	}

	// Comportamento originale: stampa lo stato della risposta
	return utils.PrintResponseState(respBody)
}
