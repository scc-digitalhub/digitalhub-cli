// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package sdk

import (
	"context"
	"fmt"
)

type GetService struct {
	http CoreHTTP
}

func NewGetService(_ context.Context, conf Config) (*GetService, error) {
	return &GetService{http: newHTTPCore(nil, conf.Core)}, nil
}

// Ritorna il body raw come prima, più status code (per eventuali usi futuri)
func (s *GetService) Get(project, endpoint, id, name string) ([]byte, int, error) {
	params := map[string]string{}
	if id == "" {
		if name == "" {
			return nil, 0, fmt.Errorf("you must specify id or name")
		}
		params["name"] = name
		params["versions"] = "latest"
	}

	url := s.http.BuildURL(project, endpoint, id, params)
	return s.http.Do(context.Background(), "GET", url, nil)
}
