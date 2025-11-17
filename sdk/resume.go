// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package sdk

import (
	"context"
	"errors"
	"fmt"
)

type ResumeRequest struct {
	Project  string
	Endpoint string // translated resource endpoint, e.g. "runs"
	ID       string
}

type ResumeService struct {
	http CoreHTTP
}

func NewResumeService(ctx context.Context, conf Config) (*ResumeService, error) {
	if conf.Core.BaseURL == "" || conf.Core.APIVersion == "" {
		return nil, errors.New("invalid core config")
	}
	return &ResumeService{
		http: newHTTPCore(nil, conf.Core),
	}, nil
}

// Resume performs POST {base}/{project}/{endpoint}/{id}/resume
// Returns response body and status code so the adapter can print state as before.
func (s *ResumeService) Resume(ctx context.Context, req ResumeRequest) ([]byte, int, error) {
	if req.Project == "" {
		return nil, 0, errors.New("project not specified")
	}
	if req.Endpoint == "" {
		return nil, 0, errors.New("endpoint not specified")
	}
	if req.ID == "" {
		return nil, 0, errors.New("id not specified")
	}

	url := s.http.BuildURL(req.Project, req.Endpoint, req.ID, nil) + "/resume"
	b, status, err := s.http.Do(ctx, "POST", url, nil)
	if err != nil {
		return nil, status, fmt.Errorf("resume request failed (status %d): %w", status, err)
	}
	return b, status, nil
}
