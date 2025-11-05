// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"sigs.k8s.io/yaml"
)

type CreateService struct {
	http CoreHTTP
}

type CreateRequest struct {
	Project  string
	Endpoint string
	Name     string
	FilePath string
	ResetID  bool
}

func NewCreateService(_ context.Context, conf Config) (*CreateService, error) {
	return &CreateService{http: newHTTPCore(nil, conf.Core)}, nil
}

func (s *CreateService) Create(req CreateRequest) error {
	var jsonMap map[string]any

	if req.FilePath != "" {
		// leggi YAML e converti in JSON -> map
		data, err := os.ReadFile(req.FilePath)
		if err != nil {
			return fmt.Errorf("failed to read YAML file: %w", err)
		}
		jsonBytes, err := yaml.YAMLToJSON(data)
		if err != nil {
			return fmt.Errorf("yaml to json failed: %w", err)
		}
		if err := json.Unmarshal(jsonBytes, &jsonMap); err != nil {
			return fmt.Errorf("failed to parse after JSON conversion: %w", err)
		}

		delete(jsonMap, "user")
		if req.Endpoint != "projects" {
			jsonMap["project"] = req.Project
		}
		if req.ResetID {
			delete(jsonMap, "id")
		}
	} else {
		// caso project senza file: usa solo name
		jsonMap = map[string]any{
			"name": req.Name,
		}
	}

	body, err := json.Marshal(jsonMap)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	// POST su /api/{ver}/[ -/{project} ]/{endpoint}
	url := s.http.BuildURL(req.Project, req.Endpoint, "", nil)
	_, _, err = s.http.Do(context.Background(), "POST", url, body)
	if err != nil {
		return err
	}
	return nil
}
