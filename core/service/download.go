// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"bytes"
	"context"
	s3client "dhcli/configs"
	"dhcli/utils"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	"sigs.k8s.io/yaml"
)

type DownloadInfo struct {
	Filename string `json:"filename" yaml:"filename"`
	Size     int64  `json:"size"     yaml:"size"`
	Path     string `json:"path"     yaml:"path"`
}

// DownloadHandler downloads artifacts and reports local target paths.
// - short: prints local paths
// - json/yaml: prints filename, size, path for each downloaded file
func DownloadHandler(env string, destination string, output string, project string, name string, resource string, id string, verbose bool) error {
	endpoint := utils.TranslateEndpoint(resource)

	if endpoint != "projects" && project == "" {
		return errors.New("project is mandatory when performing this operation on resources other than projects")
	}

	format := utils.TranslateFormat(output)

	params := map[string]string{}
	if id == "" {
		if name == "" {
			return errors.New("you must specify id or name")
		}
		params["name"] = name
		params["versions"] = "latest"
	}

	url := utils.BuildCoreUrl(project, endpoint, id, params)

	req := utils.PrepareRequest("GET", url, nil, viper.GetString(utils.DhCoreAccessToken))
	body, err := utils.DoRequest(req)
	if err != nil {
		return fmt.Errorf("error reading response: %w", err)
	}

	// Single object or list in "content"
	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return fmt.Errorf("error unmarshalling JSON: %w", err)
	}

	var contentList []interface{}
	if id != "" {
		contentList = []interface{}{raw}
	} else {
		var ok bool
		contentList, ok = raw["content"].([]interface{})
		if !ok {
			return fmt.Errorf("missing or invalid 'content' field in response")
		}
	}
	if len(contentList) == 0 {
		return fmt.Errorf("no artifact was found in Content response")
	}

	ctx := context.Background()
	var s3Client *s3client.Client

	var downloadInfos []DownloadInfo

	for i, item := range contentList {
		artifactMap, ok := item.(map[string]interface{})
		if !ok {
			log.Printf("Skipping invalid artifact at index %d", i)
			continue
		}

		spec, ok := artifactMap["spec"].(map[string]interface{})
		if !ok {
			log.Printf("Skipping artifact with missing spec field at index %d", i)
			continue
		}

		pathStr, _ := spec["path"].(string)
		if pathStr == "" {
			log.Printf("Skipping artifact with empty path at index %d", i)
			continue
		}

		parsedPath, err := utils.ParsePath(pathStr)
		if err != nil {
			return fmt.Errorf("failed to parse path: %w", err)
		}

		localFilename := parsedPath.Filename
		localPath := localFilename

		// Destination directory or file path
		if destination != "" {
			info, err := os.Stat(destination)
			if err != nil {
				if os.IsNotExist(err) {
					if err := os.MkdirAll(destination, 0o755); err != nil {
						return fmt.Errorf("failed to create destination directory: %w", err)
					}
					info, err = os.Stat(destination)
					if err != nil {
						return fmt.Errorf("failed to stat created directory: %w", err)
					}
				} else {
					return fmt.Errorf("error accessing destination path: %w", err)
				}
			}
			if info.IsDir() {
				localPath = filepath.Join(destination, localFilename)
			} else {
				localPath = destination
			}
		}

		switch parsedPath.Scheme {
		case "s3":
			if s3Client == nil {
				cfg := s3client.Config{
					AccessKey:   viper.GetString("aws_access_key_id"),
					SecretKey:   viper.GetString("aws_secret_access_key"),
					AccessToken: viper.GetString("aws_session_token"),
					Region:      viper.GetString("aws_region"),
					EndpointURL: viper.GetString("aws_endpoint_url"),
				}
				client, err := s3client.NewClient(ctx, cfg)
				if err != nil {
					return fmt.Errorf("failed to create S3 client: %w", err)
				}
				s3Client = client
			}

			// Normalizza l'S3 key (rimuove eventuale leading "/")
			parsedPath.Path = strings.TrimPrefix(parsedPath.Path, "/")

			if strings.HasSuffix(parsedPath.Path, "/") {
				// Directory download (paginato)
				if err := utils.DownloadS3FileOrDir(s3Client, ctx, parsedPath, localPath, verbose); err != nil {
					log.Println("Error downloading from S3:", err)
				}

				// Rebuild local target paths per reporting (lista completa)
				baseDir := dirBaseForLocalTarget(localPath)
				files, err := s3Client.ListFilesAll(ctx, parsedPath.Host, parsedPath.Path)
				if err != nil {
					log.Printf("Warning: failed to list S3 folder for reporting (%v)\n", err)
					break
				}
				for _, f := range files {
					relative := strings.TrimPrefix(f.Path, parsedPath.Path)
					targetPath := filepath.Join(baseDir, relative)
					if st, err := os.Stat(targetPath); err == nil && !st.IsDir() {
						downloadInfos = append(downloadInfos, DownloadInfo{
							Filename: filepath.Base(targetPath),
							Size:     st.Size(),
							Path:     targetPath,
						})
					}
				}
			} else {
				// Single file
				if err := utils.DownloadS3FileOrDir(s3Client, ctx, parsedPath, localPath, verbose); err != nil {
					log.Println("Error downloading from S3:", err)
					continue
				}
				if st, err := os.Stat(localPath); err == nil && !st.IsDir() {
					downloadInfos = append(downloadInfos, DownloadInfo{
						Filename: filepath.Base(localPath),
						Size:     st.Size(),
						Path:     localPath,
					})
				}
			}

		case "http", "https":
			if err := utils.DownloadHTTPFile(parsedPath.Path, localPath); err != nil {
				log.Println("Error downloading from HTTP/s:", err)
				continue
			}
			if st, err := os.Stat(localPath); err == nil && !st.IsDir() {
				downloadInfos = append(downloadInfos, DownloadInfo{
					Filename: filepath.Base(localPath),
					Size:     st.Size(),
					Path:     localPath,
				})
			}

		case "other", "":
			fmt.Printf("Skipping unsupported scheme: %s\n", parsedPath.Path)
			continue

		default:
			return fmt.Errorf("unsupported scheme: %s", parsedPath.Scheme)
		}
	}

	switch format {
	case "short":
		printDownloadShort(downloadInfos)
	case "json":
		if err := printDownloadJSON(downloadInfos); err != nil {
			return err
		}
	case "yaml":
		if err := printDownloadYAML(downloadInfos); err != nil {
			return err
		}
	default:
		printDownloadShort(downloadInfos)
	}

	return nil
}

// Returns the local base directory used for S3 directory downloads.
// If localPath has a single segment, base is "" (cwd); otherwise it's the parent directory.
func dirBaseForLocalTarget(localPath string) string {
	clean := filepath.Clean(localPath)
	parent := filepath.Dir(clean)
	if parent == "." || parent == string(os.PathSeparator) {
		return ""
	}
	return parent
}

func printDownloadShort(items []DownloadInfo) {
	for _, it := range items {
		fmt.Println(it.Path)
	}
}

func printDownloadJSON(items []DownloadInfo) error {
	data, err := json.Marshal(items)
	if err != nil {
		return err
	}
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, data, "", "    "); err != nil {
		return err
	}
	fmt.Println(pretty.String())
	return nil
}

func printDownloadYAML(items []DownloadInfo) error {
	out, err := yaml.Marshal(items)
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}
