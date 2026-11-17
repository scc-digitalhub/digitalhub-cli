// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"dhcli/handlers/utils"
	"dhcli/keys"

	"github.com/scc-digitalhub/digitalhub-cli-sdk/sdk/config"
	crudsvc "github.com/scc-digitalhub/digitalhub-cli-sdk/sdk/services/crud"
	runsvc "github.com/scc-digitalhub/digitalhub-cli-sdk/sdk/services/run"
	"github.com/scc-digitalhub/digitalhub-cli-sdk/sdk/services/transfer"
	"github.com/spf13/viper"
)

// GetResourceData fetches a single resource (by ID or latest version by Name) and returns both the parsed map and raw JSON bytes.
func GetResourceData(ctx context.Context, env, project, resource, name, id string) (map[string]any, []byte, error) {
	endpoint, err := utils.ResolveEndpoint(resource)
	if err != nil {
		return nil, nil, err
	}

	if err := prepareEnvironment(ctx, env, keys.GetMin, keys.GetMax); err != nil {
		return nil, nil, err
	}

	if endpoint != "projects" && project == "" {
		return nil, nil, errors.New("project is mandatory when performing this operation on resources other than projects")
	}

	cfg := NewCoreConfig()
	crud, err := crudsvc.NewCrudService(ctx, cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("sdk init failed: %w", err)
	}

	body, _, err := crud.Get(ctx, crudsvc.GetRequest{
		ResourceRequest: crudsvc.ResourceRequest{
			Project:  project,
			Resource: endpoint,
		},
		ID:   id,
		Name: url.QueryEscape(name),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("error in request: %w", err)
	}

	var m map[string]any
	if err := json.Unmarshal(body, &m); err != nil {
		return nil, nil, err
	}

	if id == "" {
		first, err := utils.GetFirstIfList(m)
		if err != nil {
			return nil, nil, err
		}
		out, err := json.Marshal(first)
		if err != nil {
			return nil, nil, err
		}
		return first, out, nil
	}

	return m, body, nil
}

// ListResourcesData lists all pages of a resource type matching optional filters.
func ListResourcesData(ctx context.Context, env, project, resource, name, kind, state string) ([]any, error) {
	endpoint, err := utils.ResolveEndpoint(resource)
	if err != nil {
		return nil, err
	}

	if err := prepareEnvironment(ctx, env, keys.ListMin, keys.ListMax); err != nil {
		return nil, err
	}

	if endpoint != "projects" && project == "" {
		return nil, errors.New("project is mandatory when performing this operation on resources other than projects")
	}

	cfg := NewCoreConfig()
	crud, err := crudsvc.NewCrudService(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("sdk init failed: %w", err)
	}

	params := map[string]string{
		"name":  name,
		"kind":  kind,
		"state": state,
		"size":  "200",
		"sort":  "updated,asc",
	}
	if name != "" {
		params["versions"] = "all"
	}
	escapeQueryValues(params)

	elements, _, err := crud.ListAllPages(ctx, crudsvc.ListRequest{
		ResourceRequest: crudsvc.ResourceRequest{
			Project:  project,
			Resource: endpoint,
		},
		Params: params,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch list: %w", err)
	}
	if elements == nil {
		elements = []any{}
	}
	return elements, nil
}

// ListServicesData lists runs with action=serve.
func ListServicesData(ctx context.Context, env, project, name, kind, state string) ([]any, error) {
	if err := prepareEnvironment(ctx, env, keys.ListMin, keys.ListMax); err != nil {
		return nil, err
	}

	if project == "" {
		return nil, errors.New("project is mandatory when listing services")
	}

	cfg := NewCoreConfig()
	crud, err := crudsvc.NewCrudService(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("sdk init failed: %w", err)
	}

	if state == "" {
		state = "RUNNING"
	}
	params := map[string]string{
		"name":   name,
		"kind":   kind,
		"state":  state,
		"action": "serve",
		"size":   "200",
		"sort":   "updated,asc",
	}
	if name != "" {
		params["versions"] = "all"
	}
	escapeQueryValues(params)

	elements, _, err := crud.ListAllPages(ctx, crudsvc.ListRequest{
		ResourceRequest: crudsvc.ResourceRequest{
			Project:  project,
			Resource: "runs",
		},
		Params: params,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch services list: %w", err)
	}
	if elements == nil {
		elements = []any{}
	}
	return elements, nil
}

func escapeQueryValues(params map[string]string) {
	for key, value := range params {
		params[key] = url.QueryEscape(value)
	}
}

// GetRunLogData fetches the decoded log text for a run, falling back to persisted project logs if live container logs are unavailable.
func GetRunLogData(ctx context.Context, env, project, id, container string) (string, error) {
	if err := prepareEnvironment(ctx, env, keys.LogMin, keys.LogMax); err != nil {
		return "", err
	}
	if project == "" {
		return "", errors.New("project not specified")
	}
	if id == "" {
		return "", errors.New("run id not specified")
	}

	cfg := NewCoreConfig()
	svc, err := runsvc.NewRunService(ctx, cfg)
	if err != nil {
		return "", err
	}

	containerLog, err := getContainerLogWithFallback(ctx, svc, project, id, container)
	if err != nil {
		return "", err
	}

	rawContent, ok := containerLog["content"].(string)
	if !ok {
		return "", errors.New("invalid log entry: missing or invalid content field")
	}

	logContents, err := base64.StdEncoding.DecodeString(rawContent)
	if err != nil {
		return "", err
	}
	return string(logContents), nil
}

func getContainerLogWithFallback(
	ctx context.Context,
	svc *runsvc.RunService,
	project string,
	id string,
	container string,
) (map[string]any, error) {
	entry, err := getContainerLogAdapter(ctx, svc, project, "runs", id, container)
	if err == nil {
		return entry, nil
	}

	// Fallback: search persisted logs in the project matching this run ID
	coreHTTP := NewCoreHTTP()
	for page := 0; ; page++ {
		params := map[string]string{"size": "200", "sort": "updated,desc"}
		if page > 0 {
			params["page"] = strconv.Itoa(page)
		}
		url := coreHTTP.BuildURL(project, "logs", "", params)
		b, _, listErr := coreHTTP.Do(ctx, "GET", url, nil)
		if listErr != nil {
			return nil, errors.Join(err, fmt.Errorf("fetch persisted logs: %w", listErr))
		}
		var parsed struct {
			Content    []map[string]any `json:"content"`
			TotalPages *int             `json:"totalPages"`
		}
		if parseErr := json.Unmarshal(b, &parsed); parseErr != nil {
			return nil, errors.Join(err, fmt.Errorf("parse persisted logs: %w", parseErr))
		}
		if parsed.Content == nil {
			return nil, errors.Join(err, errors.New("invalid persisted log response: missing content array"))
		}
		for _, m := range parsed.Content {
			if matchesRunLog(m, id, container) {
				if _, hasContent := m["content"].(string); hasContent {
					return m, nil
				}
			}
		}
		if parsed.TotalPages == nil || page+1 >= *parsed.TotalPages {
			break
		}
	}
	return nil, err
}

func matchesRunLog(entry map[string]any, runID, container string) bool {
	matchedRun := false
	if r, ok := entry["run"].(string); ok && (r == runID || strings.HasSuffix(r, "/"+runID)) {
		matchedRun = true
	}
	if spec, ok := entry["spec"].(map[string]any); ok {
		if r, ok := spec["run"].(string); ok && (r == runID || strings.HasSuffix(r, "/"+runID)) {
			matchedRun = true
		}
	}
	if !matchedRun {
		return false
	}
	if container == "" {
		return true
	}
	if idStr, _ := entry["id"].(string); idStr == container {
		return true
	}
	if status, ok := entry["status"].(map[string]any); ok {
		if cName, _ := status["container"].(string); cName == container {
			return true
		}
	}
	return false
}

// GetMetricsData fetches metrics JSON for scope ("instance", "project", "run").
func GetMetricsData(ctx context.Context, env, project, scope, id string) ([]byte, error) {
	if err := prepareEnvironment(ctx, env, keys.MetricsMin, keys.MetricsMax); err != nil {
		return nil, err
	}

	baseURL := strings.TrimRight(viper.GetString(keys.DhCoreEndpoint), "/")
	apiVersion := viper.GetString(keys.DhCoreApiVersion)

	if scope == "" {
		if id != "" {
			scope = "run"
		} else if project != "" {
			scope = "project"
		} else {
			scope = "instance"
		}
	}

	var metricsURL string
	switch scope {
	case "instance":
		metricsURL = fmt.Sprintf("%s/api/%s/resource_metrics", baseURL, apiVersion)
	case "project":
		if project == "" {
			return nil, errors.New("project name is required")
		}
		metricsURL = fmt.Sprintf("%s/api/%s/projects/%s/resource_metrics", baseURL, apiVersion, project)
	case "run":
		if project == "" {
			return nil, errors.New("project is required for run metrics")
		}
		if id == "" {
			return nil, errors.New("run id is required")
		}
		metricsURL = fmt.Sprintf("%s/api/%s/-/%s/runs/%s/resource_metrics", baseURL, apiVersion, project, id)
	default:
		return nil, fmt.Errorf("unknown scope: %s", scope)
	}

	httpClient := utils.GetDebugHTTPClient()
	if httpClient == nil {
		httpClient = &http.Client{}
	}

	return fetchMetrics(ctx, httpClient, metricsURL)
}

// StopRunData stops a running execution and returns the raw JSON response from Core.
func StopRunData(ctx context.Context, env, project, id string) ([]byte, error) {
	if err := prepareEnvironment(ctx, env, keys.StopMin, keys.StopMax); err != nil {
		return nil, err
	}
	if project == "" {
		return nil, errors.New("project not specified")
	}
	if id == "" {
		return nil, errors.New("run id not specified")
	}

	cfg := NewCoreConfig()
	svc, err := runsvc.NewRunService(ctx, cfg)
	if err != nil {
		return nil, err
	}

	respBody, _, err := svc.Stop(ctx, runsvc.StopRequest{
		RunResourceRequest: runsvc.RunResourceRequest{
			Project:  project,
			Resource: "runs",
			ID:       id,
		},
	})
	if err != nil {
		return nil, err
	}
	return respBody, nil
}

// UploadResourceData uploads a file or directory to S3 and creates/updates the corresponding entity.
func UploadResourceData(ctx context.Context, env, input, project, resource, id, name string, verbose bool) (any, error) {
	if err := prepareEnvironment(ctx, env, keys.LoginMin, keys.LoginMax); err != nil {
		return nil, err
	}
	if input == "" {
		return nil, errors.New("missing required input file or directory")
	}
	endpoint, err := utils.ResolveEndpoint(resource)
	if err != nil {
		return nil, err
	}
	if endpoint != "projects" && project == "" {
		return nil, errors.New("project is mandatory for non-project resources")
	}

	cfg := config.Config{
		Core: NewCoreConfig().Core,
		S3: config.S3Config{
			AccessKey:   viper.GetString("aws_access_key_id"),
			SecretKey:   viper.GetString("aws_secret_access_key"),
			AccessToken: viper.GetString("aws_session_token"),
			Region:      viper.GetString("aws_region"),
			EndpointURL: viper.GetString("aws_endpoint_url"),
		},
		HTTPClient: utils.GetDebugHTTPClient(),
	}

	bucket := viper.GetString("s3_bucket")
	if bucket == "" {
		bucket = "datalake"
	}

	svc, err := transfer.NewTransferService(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("sdk init failed: %w", err)
	}

	singular := strings.TrimSuffix(endpoint, "s")
	res, err := svc.Upload(ctx, endpoint, transfer.UploadRequest{
		Project:  project,
		Resource: singular,
		ID:       id,
		Name:     name,
		Input:    input,
		Verbose:  verbose,
		Bucket:   bucket,
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}
