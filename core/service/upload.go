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
	"log"
	"os"
	"strings"

	"github.com/spf13/viper"
)

func UploadHandler(env, input, project, resource string, id string, name string) error {
	if input == "" {
		return errors.New("missing required input file or directory")
	}

	endpoint := utils.TranslateEndpoint(resource)
	if endpoint != "projects" && project == "" {
		return errors.New("project is mandatory for non-project resources")
	}

	// If no ID is provided, generate a new one and then create the artifact
	if id == "" {
		// name is mandatory for new artifacts
		if name == "" {
			return errors.New("name is required when creating a new artifact")
		}

		id = utils.UUIDv4NoDash()
		log.Printf("No ID provided, generating new artifact id: %s", id)

		fileInfo, err := os.Stat(input)
		if err != nil {
			return fmt.Errorf("cannot access input: %w", err)
		}

		var path string
		if fileInfo.IsDir() {
			path = fmt.Sprintf("s3://%s/%s/%s/%s/", "datalake", project, resource, id)
		} else {
			path = fmt.Sprintf("s3://%s/%s/%s/%s/%s", "datalake", project, resource, id, fileInfo.Name())
		}

		log.Printf("S3 path for new artifact: %s", path)

		spec := map[string]interface{}{
			"path": path,
		}
		entity := map[string]interface{}{
			"id":      id,
			"project": project,
			"kind":    resource,
			"name":    name,
			"spec":    spec,
		}
		// Create the artifact with initial status
		entity["status"] = map[string]interface{}{
			"state": "CREATED",
		}

		payload, err := json.Marshal(entity)
		fmt.Printf("Payload for artifact creation: %s\n", payload)
		if err != nil {
			return fmt.Errorf("failed to marshal artifact creation payload: %w", err)
		}
		url := utils.BuildCoreUrl(project, endpoint, "", nil)
		log.Printf("Creating artifact at URL: %s", url)
		req := utils.PrepareRequest("POST", url, payload, viper.GetString(utils.DhCoreAccessToken))
		_, err = utils.DoRequest(req)
		if err != nil {
			return fmt.Errorf("failed to create artifact: %w", err)
		}
		log.Printf("Artifact created with ID: %s", id)

	}

	// From here on, we assume the artifact already exists and we are uploading to it
	url := utils.BuildCoreUrl(project, endpoint, id, nil)
	log.Printf("Requesting artifact info from URL: %s", url)

	req := utils.PrepareRequest("GET", url, nil, viper.GetString(utils.DhCoreAccessToken))
	body, err := utils.DoRequest(req)
	if err != nil {
		return fmt.Errorf("failed to retrieve artifact info: %w", err)
	}

	// Unmarshal raw JSON
	var artifactMap map[string]interface{}
	if err := json.Unmarshal(body, &artifactMap); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	// Check artifact status
	status, ok := artifactMap["status"].(map[string]interface{})
	if !ok {
		return errors.New("missing or invalid status field")
	}
	state, _ := status["state"].(string)
	if state != "CREATED" {
		return fmt.Errorf("artifact is not in CREATED state, current state: %s", state)
	}

	// Parse S3 path
	spec, ok := artifactMap["spec"].(map[string]interface{})
	if !ok {
		return errors.New("missing or invalid spec field")
	}
	pathStr, _ := spec["path"].(string)
	parsedPath, err := utils.ParsePath(pathStr)
	if err != nil {
		return fmt.Errorf("invalid path in artifact: %w", err)
	}
	if parsedPath.Scheme != "s3" {
		return fmt.Errorf("only s3 scheme is supported for upload")
	}

	// Build S3 client
	cfg := s3client.Config{
		AccessKey:   viper.GetString("aws_access_key_id"),
		SecretKey:   viper.GetString("aws_secret_access_key"),
		AccessToken: viper.GetString("aws_session_token"),
		Region:      viper.GetString("aws_region"),
		EndpointURL: viper.GetString("aws_endpoint_url"),
	}
	ctx := context.Background()
	client, err := s3client.NewClient(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to create S3 client: %w", err)
	}

	// Inner function to update the status
	updateStatus := func(artifactKey string, updateData map[string]interface{}) error {
		existingData, ok := artifactMap[artifactKey].(map[string]interface{})
		if !ok {
			existingData = make(map[string]interface{})
		}

		merged := utils.MergeMaps(existingData, updateData, utils.MergeConfig{})
		artifactMap[artifactKey] = merged

		payload, err := json.Marshal(artifactMap)
		if err != nil {
			return fmt.Errorf("failed to marshal updated artifact: %w", err)
		}

		updateURL := utils.BuildCoreUrl(project, endpoint, id, nil)
		req := utils.PrepareRequest("PUT", updateURL, payload, viper.GetString(utils.DhCoreAccessToken))

		_, err = utils.DoRequest(req)
		if err != nil {
			return fmt.Errorf("failed to update artifact status with data %v: %w", updateData, err)
		}
		return nil
	}

	// Set status to UPLOADING
	if err := updateStatus("status", map[string]interface{}{"state": "UPLOADING"}); err != nil {
		return err
	}

	// Upload
	fileInfo, err := os.Stat(input)
	if err != nil {
		_ = updateStatus("status", map[string]interface{}{"state": "ERROR"})
		return fmt.Errorf("cannot access input: %w", err)
	}

	files := make([]map[string]interface{}, 0)

	if fileInfo.IsDir() {
		_, files, err = utils.UploadS3Dir(client, ctx, parsedPath, input)
		if err != nil {
			_ = updateStatus("status", map[string]interface{}{"state": "ERROR"})
			return fmt.Errorf("upload failed: %w", err)
		}
	} else {
		var targetKey string
		if strings.HasSuffix(parsedPath.Path, "/") {
			targetKey = parsedPath.Path + fileInfo.Name()
		} else {
			targetKey = parsedPath.Path
		}
		_, files, err = utils.UploadS3File(client, ctx, parsedPath.Host, targetKey, input)
		if err != nil {
			log.Fatalf("Upload failed: %v", err)
		}

	}

	if err := updateStatus("status", map[string]interface{}{
		"state": "READY",
		"files": files,
	}); err != nil {
		return fmt.Errorf("upload succeeded but failed to update status: %w", err)
	}

	log.Println("Upload successful and artifact status updated to READY.")
	return nil
}
