// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package sdk

import (
	"context"
	s3client "dhcli/configs"
	"dhcli/utils"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// UploadService: estrazione 1:1 della semantica dell'UploadHandler originale.
type UploadService struct {
	http CoreHTTP
	s3   *s3client.Client
	cfg  Config
}

func NewUploadService(ctx context.Context, conf Config) (*UploadService, error) {
	httpc := newHTTPCore(nil, conf.Core)

	s3c, err := s3client.NewClient(ctx, s3client.Config{
		AccessKey:   conf.S3.AccessKey,
		SecretKey:   conf.S3.SecretKey,
		AccessToken: conf.S3.SessionToken,
		Region:      conf.S3.Region,
		EndpointURL: conf.S3.EndpointURL,
	})
	if err != nil {
		return nil, fmt.Errorf("S3 init failed: %w", err)
	}
	return &UploadService{http: httpc, s3: s3c, cfg: conf}, nil
}

// Upload esegue:
// - creazione artefatto (se ID vuoto) in stato CREATED con spec.path su S3
// - transizione a UPLOADING
// - upload file/dir verso s3://<bucket>/<project>/<resource>/<id>/...
// - transizione a READY con files[] allegati
func (s *UploadService) Upload(ctx context.Context, endpoint string, req UploadRequest) (*UploadResult, error) {
	if req.Input == "" {
		return nil, errors.New("missing required input file or directory")
	}
	if endpoint != "projects" && req.Project == "" {
		return nil, errors.New("project is mandatory for non-project resources")
	}

	// 1) Se ID vuoto: creare l'artefatto
	artifactID := req.ID
	if artifactID == "" {
		if req.Name == "" {
			return nil, errors.New("name is required when creating a new artifact")
		}
		bucket := req.Bucket
		if bucket == "" {
			bucket = "datalake" // retro-compat
		}

		st, err := os.Stat(req.Input)
		if err != nil {
			return nil, fmt.Errorf("cannot access input: %w", err)
		}

		artifactID = utils.UUIDv4NoDash()

		var path string
		if st.IsDir() {
			path = fmt.Sprintf("s3://%s/%s/%s/%s/", bucket, req.Project, req.Resource, artifactID)
		} else {
			path = fmt.Sprintf("s3://%s/%s/%s/%s/%s", bucket, req.Project, req.Resource, artifactID, st.Name())
		}

		entity := map[string]interface{}{
			"id":      artifactID,
			"project": req.Project,
			"kind":    req.Resource,
			"name":    req.Name,
			"spec": map[string]interface{}{
				"path": path,
			},
			"status": map[string]interface{}{
				"state": "CREATED",
			},
		}
		payload, err := json.Marshal(entity)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal artifact creation payload: %w", err)
		}

		createURL := s.http.BuildURL(req.Project, endpoint, "", nil)

		if _, _, err = s.http.Do(ctx, "POST", createURL, payload); err != nil {
			return nil, fmt.Errorf("failed to create artifact: %w", err)
		}
	}

	// 2) Recupera l'artefatto
	getURL := s.http.BuildURL(req.Project, endpoint, artifactID, nil)
	body, _, err := s.http.Do(ctx, "GET", getURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve artifact info: %w", err)
	}
	var artifact map[string]interface{}
	if err := json.Unmarshal(body, &artifact); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// 3) Verifica stato
	status, ok := artifact["status"].(map[string]interface{})
	if !ok {
		return nil, errors.New("missing or invalid status field")
	}
	state, _ := status["state"].(string)
	if state != "CREATED" {
		return nil, fmt.Errorf("artifact is not in CREATED state, current state: %s", state)
	}

	// 4) Parse spec.path (deve essere s3)
	spec, ok := artifact["spec"].(map[string]interface{})
	if !ok {
		return nil, errors.New("missing or invalid spec field")
	}
	pathStr, _ := spec["path"].(string)
	parsedPath, err := utils.ParsePath(pathStr)
	if err != nil {
		return nil, fmt.Errorf("invalid path in artifact: %w", err)
	}
	if parsedPath.Scheme != "s3" {
		return nil, fmt.Errorf("only s3 scheme is supported for upload")
	}

	// 5) Helper: update status sul Core (merge preservando altri campi)
	updateStatus := func(key string, updateData map[string]interface{}) error {
		existing, ok := artifact[key].(map[string]interface{})
		if !ok {
			existing = map[string]interface{}{}
		}
		merged := utils.MergeMaps(existing, updateData, utils.MergeConfig{})
		artifact[key] = merged

		payload, err := json.Marshal(artifact)
		if err != nil {
			return fmt.Errorf("failed to marshal updated artifact: %w", err)
		}
		putURL := s.http.BuildURL(req.Project, endpoint, artifactID, nil)
		if _, _, err = s.http.Do(ctx, "PUT", putURL, payload); err != nil {
			return fmt.Errorf("failed to update artifact status with data %v: %w", updateData, err)
		}
		return nil
	}

	// 6) Stato → UPLOADING
	if err := updateStatus("status", map[string]interface{}{"state": "UPLOADING"}); err != nil {
		return nil, err
	}

	// 7) Upload
	st, err := os.Stat(req.Input)
	if err != nil {
		_ = updateStatus("status", map[string]interface{}{"state": "ERROR"})
		return nil, fmt.Errorf("cannot access input: %w", err)
	}

	var files []map[string]interface{}
	ctxUp := ctx

	if st.IsDir() {
		_, files, err = utils.UploadS3Dir(s.s3, ctxUp, parsedPath, req.Input, req.Verbose)
		if err != nil {
			_ = updateStatus("status", map[string]interface{}{"state": "ERROR"})
			return nil, fmt.Errorf("upload failed: %w", err)
		}
	} else {
		var targetKey string
		if strings.HasSuffix(parsedPath.Path, "/") {
			targetKey = filepath.ToSlash(filepath.Join(parsedPath.Path, st.Name()))
		} else {
			targetKey = parsedPath.Path
		}
		_, files, err = utils.UploadS3File(s.s3, ctxUp, parsedPath.Host, targetKey, req.Input, req.Verbose)
		if err != nil {
			_ = updateStatus("status", map[string]interface{}{"state": "ERROR"})
			return nil, fmt.Errorf("upload failed: %w", err)
		}
	}

	// 8) Stato → READY + files
	if err := updateStatus("status", map[string]interface{}{
		"state": "READY",
		"files": files,
	}); err != nil {
		return &UploadResult{ArtifactID: artifactID, Files: files}, fmt.Errorf("upload succeeded but failed to update status: %w", err)
	}

	return &UploadResult{ArtifactID: artifactID, Files: files}, nil
}
