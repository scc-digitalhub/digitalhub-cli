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

// Service di download focalizzato sugli artifacts/path remoti (S3/HTTP)
type DownloadService struct {
	http CoreHTTP
	s3   *s3client.Client
}

func NewDownloadService(ctx context.Context, conf Config) (*DownloadService, error) {
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

	return &DownloadService{http: httpc, s3: s3c}, nil
}

func (s *DownloadService) Download(ctx context.Context, req DownloadRequest) ([]DownloadInfo, error) {
	if req.Resource != "projects" && req.Project == "" {
		return nil, errors.New("project is mandatory for non-project resources")
	}
	if req.ID == "" && req.Name == "" {
		return nil, errors.New("you must specify id or name")
	}

	params := map[string]string{}
	id := req.ID
	if id == "" {
		params["name"] = req.Name
		params["versions"] = "latest"
	}

	url := s.http.BuildURL(req.Project, req.Resource, id, params)
	body, _, err := s.http.Do(ctx, "GET", url, nil)
	if err != nil {
		// httpcore.Do già restituisce errore formattato come l’originale
		return nil, err
	}

	paths, err := extractPaths(body)
	if err != nil {
		return nil, err
	}

	var out []DownloadInfo
	for _, p := range paths {
		pp, err := utils.ParsePath(p)
		if err != nil {
			// comportamento originale: skip elemento malformato
			continue
		}
		target, createdDir, err := chooseLocalTarget(req.Destination, pp.Filename)
		if err != nil {
			// skip come l’originale in caso di path locale non creabile
			continue
		}
		_ = createdDir // non serve usarlo ora, ma è allineato al comportamento di creazione

		switch pp.Scheme {
		case "s3":
			key := strings.TrimPrefix(pp.Path, "/")
			if strings.HasSuffix(key, "/") {
				// Directory (paginata): in caso di errore, NON fallire tutto → skip (come original)
				if derr := utils.DownloadS3FileOrDir(s.s3, ctx, pp, target, req.Verbose); derr != nil {
					// skip dir (log a livello CLI se vuoi)
					continue
				}
				// reporting
				files, lerr := s.s3.ListFilesAll(ctx, pp.Host, key)
				if lerr != nil {
					// warning/skip reporting, ma NON fallire
					continue
				}
				base := dirBaseForLocalTarget(target)
				for _, f := range files {
					local := filepath.Join(base, strings.TrimPrefix(f.Path, key))
					if st, err := os.Stat(local); err == nil && !st.IsDir() {
						out = append(out, DownloadInfo{
							Filename: filepath.Base(local),
							Size:     st.Size(),
							Path:     local,
						})
					}
				}
			} else {
				// File singolo: su errore, NON fallire → skip (come original)
				if ferr := utils.DownloadS3FileOrDir(s.s3, ctx, pp, target, req.Verbose); ferr != nil {
					continue
				}
				if st, err := os.Stat(target); err == nil && !st.IsDir() {
					out = append(out, DownloadInfo{
						Filename: filepath.Base(target),
						Size:     st.Size(),
						Path:     target,
					})
				}
			}

		case "http", "https":
			// Su errore HTTP, skip (come original)
			if herr := utils.DownloadHTTPFile(pp.Path, target); herr != nil {
				continue
			}
			if st, err := os.Stat(target); err == nil && !st.IsDir() {
				out = append(out, DownloadInfo{
					Filename: filepath.Base(target),
					Size:     st.Size(),
					Path:     target,
				})
			}

		default:
			// unsupported → skip (come original)
			continue
		}
	}
	return out, nil
}

// --- helpers ---

// chooseLocalTarget replica l’originale:
// - se dst è vuoto → usa filename nella cwd
// - se dst esiste ed è directory → dst/filename
// - se dst esiste ed è file → dst
// - se dst NON esiste → crea directory dst e usa dst/filename
func chooseLocalTarget(dst, filename string) (target string, createdDir bool, err error) {
	if dst == "" {
		return filename, false, nil
	}
	info, statErr := os.Stat(dst)
	if statErr == nil {
		if info.IsDir() {
			return filepath.Join(dst, filename), false, nil
		}
		return dst, false, nil // file esistente
	}
	if os.IsNotExist(statErr) {
		// Comportamento originale: crea directory e usa dst/filename
		if mkErr := os.MkdirAll(dst, 0o755); mkErr != nil {
			return "", false, mkErr
		}
		return filepath.Join(dst, filename), true, nil
	}
	// altro errore su Stat → propaga (ma il chiamante fa skip)
	return "", false, statErr
}

func dirBaseForLocalTarget(localPath string) string {
	clean := filepath.Clean(localPath)
	parent := filepath.Dir(clean)
	if parent == "." || parent == string(os.PathSeparator) {
		return ""
	}
	return parent
}

func extractPaths(body []byte) ([]string, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("invalid json: %w", err)
	}
	// ID specificato → singolo oggetto
	if _, has := raw["content"]; !has {
		if spec, ok := raw["spec"].(map[string]interface{}); ok {
			if path, _ := spec["path"].(string); path != "" {
				return []string{path}, nil
			}
		}
		return nil, fmt.Errorf("missing spec.path")
	}
	// content[]
	content, ok := raw["content"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid content")
	}
	var paths []string
	for _, it := range content {
		if m, ok := it.(map[string]interface{}); ok {
			if spec, ok2 := m["spec"].(map[string]interface{}); ok2 {
				if p, _ := spec["path"].(string); p != "" {
					paths = append(paths, p)
				}
			}
		}
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("no paths in content")
	}
	return paths, nil
}
