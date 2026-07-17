// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/scc-digitalhub/digitalhub-cli-sdk/sdk/config"
	crudsvc "github.com/scc-digitalhub/digitalhub-cli-sdk/sdk/services/crud"
	"github.com/spf13/viper"

	"dhcli/handlers/utils"
)

// ResolveRunIDByFunctionName finds the ID of the most recent run matching the
// given state and action for a function. The function URI is built by first
// fetching the function to retrieve its kind, then constructing
// <kind>://<project>/<functionName>.
func ResolveRunIDByFunctionName(project, functionName, state, action string) (string, error) {
	cfg := config.Config{
		Core: config.CoreConfig{
			BaseURL:     viper.GetString(utils.DhCoreEndpoint),
			APIVersion:  viper.GetString(utils.DhCoreApiVersion),
			AccessToken: viper.GetString(utils.DhCoreAccessToken),
		},
		HTTPClient: utils.GetDebugHTTPClient(),
	}

	ctx := context.Background()

	crud, err := crudsvc.NewCrudService(ctx, cfg)
	if err != nil {
		return "", fmt.Errorf("sdk init failed: %w", err)
	}

	// Fetch function to resolve its kind
	fnBody, _, err := crud.Get(ctx, crudsvc.GetRequest{
		ResourceRequest: crudsvc.ResourceRequest{
			Project:  project,
			Resource: "functions",
		},
		Name: functionName,
	})
	if err != nil {
		return "", fmt.Errorf("failed to fetch function %q: %w", functionName, err)
	}

	var fnRaw map[string]interface{}
	if err := json.Unmarshal(fnBody, &fnRaw); err != nil {
		return "", fmt.Errorf("failed to parse function response: %w", err)
	}

	// Response may be a paged list ({"content":[...]}) or a direct object
	fnMap := fnRaw
	if content, ok := fnRaw["content"].([]interface{}); ok {
		if len(content) == 0 {
			return "", fmt.Errorf("function %q not found in project %q", functionName, project)
		}
		item, ok := content[0].(map[string]interface{})
		if !ok {
			return "", fmt.Errorf("unexpected function list item format")
		}
		fnMap = item
	}

	kind, ok := fnMap["kind"].(string)
	if !ok || kind == "" {
		return "", fmt.Errorf("kind not found in function %q response", functionName)
	}

	functionURI := fmt.Sprintf("%s://%s/%s", kind, project, functionName)

	params := map[string]string{
		"function": functionURI,
		"state":    state,
		"action":   action,
		"sort":     "created,DESC",
		"size":     "1",
	}

	elements, _, err := crud.ListAllPages(ctx, crudsvc.ListRequest{
		ResourceRequest: crudsvc.ResourceRequest{
			Project:  project,
			Resource: "runs",
		},
		Params: params,
	})
	if err != nil {
		return "", fmt.Errorf("failed to fetch runs: %w", err)
	}

	if len(elements) == 0 {
		return "", fmt.Errorf("no run found for function %q (state=%s, action=%s) in project %q", functionName, state, action, project)
	}

	m, ok := elements[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("unexpected run response format")
	}

	id, ok := m["id"].(string)
	if !ok || id == "" {
		return "", fmt.Errorf("run ID not found in response")
	}

	return id, nil
}

// ResolveRunIDByName finds the ID of the most recent run matching the given
// name, state and action by querying runs?name=<name>&state=<state>&...
func ResolveRunIDByName(project, name, state, action string) (string, error) {
	cfg := config.Config{
		Core: config.CoreConfig{
			BaseURL:     viper.GetString(utils.DhCoreEndpoint),
			APIVersion:  viper.GetString(utils.DhCoreApiVersion),
			AccessToken: viper.GetString(utils.DhCoreAccessToken),
		},
		HTTPClient: utils.GetDebugHTTPClient(),
	}

	ctx := context.Background()

	crud, err := crudsvc.NewCrudService(ctx, cfg)
	if err != nil {
		return "", fmt.Errorf("sdk init failed: %w", err)
	}

	params := map[string]string{
		"name":   name,
		"state":  state,
		"action": action,
		"sort":   "created,DESC",
		"size":   "1",
	}

	elements, _, err := crud.ListAllPages(ctx, crudsvc.ListRequest{
		ResourceRequest: crudsvc.ResourceRequest{
			Project:  project,
			Resource: "runs",
		},
		Params: params,
	})
	if err != nil {
		return "", fmt.Errorf("failed to fetch runs: %w", err)
	}

	if len(elements) == 0 {
		return "", fmt.Errorf("no run found with name %q (state=%s, action=%s) in project %q", name, state, action, project)
	}

	m, ok := elements[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("unexpected run response format")
	}

	id, ok := m["id"].(string)
	if !ok || id == "" {
		return "", fmt.Errorf("run ID not found in response")
	}

	return id, nil
}
