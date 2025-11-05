// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package sdk

import (
	"context"
	"errors"
	"fmt"
)

type DeleteRequest struct {
	Project  string // obbligatorio per risorse != "projects"
	Endpoint string // chiave canonica: "projects", "artifacts", ...
	ID       string // se presente, delete di quella versione
	Name     string // se ID vuoto:
	//  - projects: nella CLI era usato per impostare ID=Name
	//  - altre risorse: delete di TUTTE le versioni (versions=all + name=...)
	Cascade bool // query param cascade=true/false
}

type DeleteService struct {
	http CoreHTTP
}

func NewDeleteService(ctx context.Context, conf Config) (*DeleteService, error) {
	if conf.Core.BaseURL == "" || conf.Core.APIVersion == "" {
		return nil, errors.New("invalid core config")
	}
	return &DeleteService{
		http: newHTTPCore(nil, conf.Core),
	}, nil
}

func (s *DeleteService) Delete(ctx context.Context, req DeleteRequest) error {
	if req.Endpoint == "" {
		return errors.New("endpoint is required")
	}

	// Validazioni come nella CLI
	if req.Endpoint != "projects" && req.Project == "" {
		return errors.New("project is mandatory for non-project resources")
	}
	if req.ID == "" && req.Name == "" {
		return errors.New("you must specify id or name")
	}

	// Query string
	params := map[string]string{
		"cascade": "false",
	}
	if req.Cascade {
		params["cascade"] = "true"
	}

	id := req.ID
	// Comportamento identico:
	//  - se id mancante e NON projects: delete all versions by name
	//  - se projects: l’ID è il nome (già impostato a livello adapter)
	if id == "" && req.Endpoint != "projects" {
		params["name"] = req.Name
		params["versions"] = "all"
	}

	url := s.http.BuildURL(req.Project, req.Endpoint, id, params)

	_, status, err := s.http.Do(ctx, "DELETE", url, nil)
	if err != nil {
		// Manteniamo un messaggio chiaro come nella CLI (la CLI già loggava
		// lo status nel DoRequest; qui lo riportiamo nell’errore)
		return fmt.Errorf("delete failed (status %d): %w", status, err)
	}
	return nil
}
