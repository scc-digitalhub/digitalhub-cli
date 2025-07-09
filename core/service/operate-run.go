// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"dhcli/utils"
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"log"
	"slices"
	"strings"
)

var validOperations = []string{"stop", "run", "resume", "delete", "build"}

func OperateRunHandler(env string, project string, id string, operation string) error {
	// Check that CLI has permission to handle runs
	endpoint := utils.TranslateEndpoint("runs")

	// Check that operation is valid
	op := strings.ToLower(operation)
	if !slices.Contains(validOperations, op) {
		return errors.New(fmt.Sprintf("Operation '%v' not supported. Supported operations: %v", op, strings.Join(validOperations, ", ")))
	}

	// Load environment and check API level requirements

	utils.CheckUpdateEnvironment()
	utils.CheckApiLevel(utils.ApiLevelKey, utils.OperateRunMin, utils.OperateRunMax)

	// Request
	method := "POST"
	url := utils.BuildCoreUrl(project, endpoint, id, nil) + "/" + op
	req := utils.PrepareRequest(method, url, nil, viper.GetString("access_token"))

	_, err := utils.DoRequest(req)
	if err != nil {
		return err
	}
	log.Println("Operation successful.")

	return nil
}
