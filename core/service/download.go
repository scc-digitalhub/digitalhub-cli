// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"
	s3client "dhcli/configs"
	"dhcli/utils"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"log"
	"os"
	"path/filepath"
)

func DownloadHandler(env string, output string, project string, name string, resource string, id string) error {
	endpoint := utils.TranslateEndpoint(resource)

	if endpoint != "projects" && project == "" {
		return errors.New("project is mandatory when performing this operation on resources other than projects")
	}

	params := map[string]string{}
	if id == "" {
		if name == "" {
			return errors.New("you must specify id or name")
		}
		params["name"] = name
		params["versions"] = "latest"
	}

	url := utils.BuildCoreUrl(project, endpoint, id, params)

	req := utils.PrepareRequest("GET", url, nil, viper.GetString("access_token"))
	body, err := utils.DoRequest(req)
	if err != nil {
		return fmt.Errorf("error reading response: %w", err)
	}

	// Parse as raw map instead of typed response
	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return fmt.Errorf("error unmarshalling JSON: %w", err)
	}

	contentList, ok := raw["content"].([]interface{})
	if !ok || len(contentList) == 0 {
		return fmt.Errorf("no artifact was found in Content response")
	}

	ctx := context.Background()
	var s3Client *s3client.Client

	for i, item := range contentList {
		artifactMap, ok := item.(map[string]interface{})
		if !ok {
			log.Printf("Skipping invalid artifact at index %d", i)
			continue
		}

		// Extract spec.path
		spec, ok := artifactMap["spec"].(map[string]interface{})
		if !ok {
			log.Printf("Skipping artifact with missing spec field at index %d", i)
			continue
		}

		pathStr, _ := spec["path"].(string)
		fmt.Printf("Entity #%d - Path: %s\n", i+1, pathStr)

		parsedPath, err := utils.ParsePath(pathStr)
		if err != nil {
			return fmt.Errorf("failed to parse path: %w", err)
		}

		localFilename := parsedPath.Filename
		localPath := localFilename

		// if output is specified, use it as the base directory if it exists or create it
		if output != "" {
			info, err := os.Stat(output)
			if err != nil {
				if os.IsNotExist(err) {
					if err := os.MkdirAll(output, 0755); err != nil {
						return fmt.Errorf("failed to create output directory: %w", err)
					}
					info, err = os.Stat(output)
					if err != nil {
						return fmt.Errorf("failed to stat created directory: %w", err)
					}
				} else {
					// Some other error (e.g., permission)
					return fmt.Errorf("error accessing output path: %w", err)
				}
			}

			if info.IsDir() {
				localPath = filepath.Join(output, localFilename)
			} else {
				localPath = output
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
			if err := utils.DownloadS3FileOrDir(s3Client, ctx, parsedPath, localPath); err != nil {
				log.Println("Error downloading from S3:", err)
			}

		case "http", "https":
			if err := utils.DownloadHTTPFile(parsedPath.Path, localPath); err != nil {
				log.Println("Error downloading from HTTP/s:", err)
			}

		case "other", "":
			fmt.Printf("Skipping unsupported scheme.....: %s\n", parsedPath.Path)

		default:
			return fmt.Errorf("unsupported scheme: %s", parsedPath.Scheme)
		}
	}

	log.Println("All files downloaded successfully.")
	return nil
}
