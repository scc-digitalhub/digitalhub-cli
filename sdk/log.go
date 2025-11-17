// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package sdk

import (
	"context"
	"errors"
	"fmt"
)

type LogRequest struct {
	Project  string
	Endpoint string // translated resource endpoint, e.g. "runs"
	ID       string
}

type LogService struct {
	http CoreHTTP
}

func NewLogService(ctx context.Context, conf Config) (*LogService, error) {
	if conf.Core.BaseURL == "" || conf.Core.APIVersion == "" {
		return nil, errors.New("invalid core config")
	}
	return &LogService{
		http: newHTTPCore(nil, conf.Core),
	}, nil
}

// GetLogs performs GET {base}/{project}/{endpoint}/{id}/logs
func (s *LogService) GetLogs(ctx context.Context, req LogRequest) ([]byte, int, error) {
	if req.Project == "" {
		return nil, 0, errors.New("project not specified")
	}
	if req.Endpoint == "" {
		return nil, 0, errors.New("endpoint not specified")
	}
	if req.ID == "" {
		return nil, 0, errors.New("id not specified")
	}

	url := s.http.BuildURL(req.Project, req.Endpoint, req.ID, nil) + "/logs"
	b, status, err := s.http.Do(ctx, "GET", url, nil)
	if err != nil {
		return nil, status, fmt.Errorf("get logs failed (status %d): %w", status, err)
	}
	return b, status, nil
}

// GetResource performs GET {base}/{project}/{endpoint}/{id}
// used to discover spec.task and compute main container name
func (s *LogService) GetResource(ctx context.Context, req LogRequest) ([]byte, int, error) {
	if req.Project == "" {
		return nil, 0, errors.New("project not specified")
	}
	if req.Endpoint == "" {
		return nil, 0, errors.New("endpoint not specified")
	}
	if req.ID == "" {
		return nil, 0, errors.New("id not specified")
	}

	url := s.http.BuildURL(req.Project, req.Endpoint, req.ID, nil)
	b, status, err := s.http.Do(ctx, "GET", url, nil)
	if err != nil {
		return nil, status, fmt.Errorf("get resource failed (status %d): %w", status, err)
	}
	return b, status, nil
}
