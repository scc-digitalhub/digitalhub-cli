// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
)

type MetricsRequest struct {
	Project   string
	Endpoint  string // result of utils.TranslateEndpoint(resource), es. "runs"
	ID        string
	Container string // optional; if empty, infer main container like original code
}

type MetricsService struct {
	http CoreHTTP
}

func NewMetricsService(ctx context.Context, conf Config) (*MetricsService, error) {
	if conf.Core.BaseURL == "" || conf.Core.APIVersion == "" {
		return nil, errors.New("invalid core config")
	}
	return &MetricsService{
		http: newHTTPCore(nil, conf.Core),
	}, nil
}

// PrintMetrics replica il comportamento di service.MetricsHandler:
// - chiama getContainerLog
// - prende status.metrics
// - se non ci sono metrics, stampa "No metrics for this run."
// - altrimenti fa pretty-print del JSON.
func (s *MetricsService) PrintMetrics(ctx context.Context, req MetricsRequest) error {
	if req.Project == "" {
		return errors.New("project not specified")
	}
	if req.Endpoint == "" {
		return errors.New("endpoint not specified")
	}
	if req.ID == "" {
		return errors.New("resource id not specified")
	}

	containerLog, err := s.getContainerLog(ctx, req.Project, req.Endpoint, req.ID, req.Container)
	if err != nil {
		return err
	}

	statusMap, ok := containerLog["status"].(map[string]interface{})
	if !ok {
		return errors.New("invalid log entry: missing status")
	}

	metricsVal := statusMap["metrics"]
	if metricsVal == nil {
		log.Println("No metrics for this run.")
		return nil
	}

	// type assertion corretta (come nel codice originale, che assume []interface{})
	metricsSlice, ok := metricsVal.([]interface{})
	if !ok {
		return errors.New("invalid metrics format")
	}

	jsonData, err := json.Marshal(metricsSlice)
	if err != nil {
		return err
	}

	var pretty bytes.Buffer
	if err := json.Indent(&pretty, jsonData, "", "    "); err != nil {
		return err
	}
	fmt.Println(pretty.String())

	return nil
}

// getContainerLog replica la logica di service.GetContainerLog:
// - GET /<endpoint>/<id>/logs
// - se container non è specificato, ricava il main container da spec.task
// - restituisce l'entry corrispondente al container scelto.
func (s *MetricsService) getContainerLog(
	ctx context.Context,
	project string,
	endpoint string,
	id string,
	container string,
) (map[string]interface{}, error) {

	// 1) GET logs
	url := s.http.BuildURL(project, endpoint, id, nil) + "/logs"
	body, status, err := s.http.Do(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("logs request failed (status %d): %w", status, err)
	}

	var logs []interface{}
	if err := json.Unmarshal(body, &logs); err != nil {
		return nil, fmt.Errorf("json parsing failed: %w", err)
	}

	// 2) Determine container name
	containerName := container
	if containerName == "" {
		// same logic as original GetContainerLog when container is not specified
		urlRes := s.http.BuildURL(project, endpoint, id, nil)
		resBody, status, err := s.http.Do(ctx, "GET", urlRes, nil)
		if err != nil {
			return nil, fmt.Errorf("resource request failed (status %d): %w", status, err)
		}

		var m map[string]interface{}
		if err := json.Unmarshal(resBody, &m); err != nil {
			return nil, err
		}

		spec, ok := m["spec"].(map[string]interface{})
		if !ok {
			return nil, errors.New("invalid resource: missing spec")
		}
		task, ok := spec["task"].(string)
		if !ok {
			return nil, errors.New("invalid resource: missing task in spec")
		}

		// Same formatting as original:
		// taskFormatted := strings.ReplaceAll(task[:strings.Index(task, ":")], "+", "")
		idx := strings.Index(task, ":")
		if idx == -1 {
			return nil, errors.New("invalid task format in spec")
		}
		taskFormatted := strings.ReplaceAll(task[:idx], "+", "")

		containerName = fmt.Sprintf("c-%v-%v", taskFormatted, id)
	}

	// 3) Find matching log entry
	for _, entry := range logs {
		entryMap, ok := entry.(map[string]interface{})
		if !ok {
			continue
		}
		statusVal, ok := entryMap["status"].(map[string]interface{})
		if !ok {
			continue
		}
		entryContainer, _ := statusVal["container"].(string)
		if entryContainer == containerName {
			return entryMap, nil
		}
	}

	return nil, fmt.Errorf("container not found")
}
