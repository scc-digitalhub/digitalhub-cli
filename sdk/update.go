// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package sdk

import (
	"context"
	"errors"
	"fmt"
)

type UpdateRequest struct {
	Project  string // obbligatorio per risorse != "projects"
	Endpoint string // "projects", "artifacts", ...
	ID       string // sempre richiesto dall'originale (update <resource> <id>)
	Body     []byte // payload JSON già pronto (mutazioni fatte nell'adapter)
}

type UpdateService struct {
	http CoreHTTP
}

func NewUpdateService(ctx context.Context, conf Config) (*UpdateService, error) {
	if conf.Core.BaseURL == "" || conf.Core.APIVersion == "" {
		return nil, errors.New("invalid core config")
	}
	return &UpdateService{
		http: newHTTPCore(nil, conf.Core),
	}, nil
}

func (s *UpdateService) Update(ctx context.Context, req UpdateRequest) error {
	// Validazioni equivalenti
	if req.Endpoint == "" {
		return errors.New("endpoint is required")
	}
	if req.ID == "" {
		return errors.New("id is required")
	}
	if req.Endpoint != "projects" && req.Project == "" {
		return errors.New("project is mandatory for non-project resources")
	}
	if len(req.Body) == 0 {
		return errors.New("empty body")
	}

	// Build URL e PUT, come faceva l’originale
	url := s.http.BuildURL(req.Project, req.Endpoint, req.ID, nil)
	_, status, err := s.http.Do(ctx, "PUT", url, req.Body)
	if err != nil {
		return fmt.Errorf("update failed (status %d): %w", status, err)
	}
	return nil
}
