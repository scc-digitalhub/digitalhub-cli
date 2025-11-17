// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package sdk

import (
	"context"
	"errors"
	"fmt"
)

type StopRequest struct {
	Project  string
	Endpoint string // already translated resource endpoint
	ID       string
}

type StopService struct {
	http CoreHTTP
}

func NewStopService(ctx context.Context, conf Config) (*StopService, error) {
	if conf.Core.BaseURL == "" || conf.Core.APIVersion == "" {
		return nil, errors.New("invalid core config")
	}
	return &StopService{
		http: newHTTPCore(nil, conf.Core),
	}, nil
}

// Stop performs POST {base}/{project}/{endpoint}/{id}/stop
// Returns response body and status code so the adapter can print the state like the legacy code.
func (s *StopService) Stop(ctx context.Context, req StopRequest) ([]byte, int, error) {
	if req.Project == "" {
		return nil, 0, errors.New("project not specified")
	}
	if req.Endpoint == "" {
		return nil, 0, errors.New("endpoint not specified")
	}
	if req.ID == "" {
		return nil, 0, errors.New("id not specified")
	}

	url := s.http.BuildURL(req.Project, req.Endpoint, req.ID, nil) + "/stop"
	// body is nil, same as legacy
	b, status, err := s.http.Do(ctx, "POST", url, nil)
	if err != nil {
		return nil, status, fmt.Errorf("stop request failed (status %d): %w", status, err)
	}
	return b, status, nil
}
