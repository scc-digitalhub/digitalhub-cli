// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

// service/download.go
package service

import (
	"context"
	s3client "dhcli/configs"
	"dhcli/models"
	"dhcli/utils"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func DownloadHandler(env string, output string, project string, name string, resource string, id string, originalArgs []string) error {
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

	_, section := utils.LoadIniConfig([]string{env})
	method := "GET"
	url := utils.BuildCoreUrl(section, project, endpoint, id, params)
	req := utils.PrepareRequest(method, url, nil, section.Key("access_token").String())
	body, err := utils.DoRequest(req)
	if err != nil {
		return fmt.Errorf("error reading response: %w", err)
	}

	var resp models.Response[models.Entity]
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("error unmarshalling JSON: %w", err)
	}
	if len(resp.Content) == 0 {
		return fmt.Errorf("no artifact was found in Content response")
	}

	ctx := context.Background()
	var s3Client *s3client.Client

	for i, artifact := range resp.Content {
		fmt.Printf("Entity #%d - Path: %s\n", i+1, artifact.Spec.Path)

		parsedPath, err := utils.ParsePath(artifact.Spec.Path)
		if err != nil {
			return fmt.Errorf("failed to parse path: %w", err)
		}

		localFilename := parsedPath.Filename
		localPath := localFilename
		if output != "" {
			info, err := os.Stat(output)
			if err != nil {
				return fmt.Errorf("output path does not exist: %s", output)
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
					AccessKey:   section.Key("aws_access_key_id").String(),
					SecretKey:   section.Key("aws_secret_access_key").String(),
					AccessToken: section.Key("aws_session_token").String(),
					Region:      section.Key("aws_region").String(),
					EndpointURL: section.Key("aws_endpoint_url").String(),
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
			fmt.Printf("Skipping other.....: %s\n", parsedPath.Path)

		default:
			return fmt.Errorf("unsupported scheme: %s", parsedPath.Scheme)
		}
	}

	log.Println("All files downloaded successfully.")
	return nil
}
