// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"bytes"
	"context"
	"dhcli/sdk/config"
	crudsvc "dhcli/sdk/services/crud"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/viper"
	"sigs.k8s.io/yaml"

	"dhcli/utils"
)

func GetHandler(env string, output string, project string, name string, resource string, id string) error {
	endpoint := utils.TranslateEndpoint(resource)

	// Stessa logica esistente
	utils.CheckUpdateEnvironment()
	utils.CheckApiLevel(utils.ApiLevelKey, utils.GetMin, utils.GetMax)

	format := utils.TranslateFormat(output)

	if endpoint != "projects" && project == "" {
		return errors.New("project is mandatory when performing this operation on resources other than projects")
	}

	// Adapter: viper/ini/env -> sdk.Config
	cfg := config.Config{
		Core: config.CoreConfig{
			BaseURL:     viper.GetString(utils.DhCoreEndpoint),
			APIVersion:  viper.GetString(utils.DhCoreApiVersion),
			AccessToken: viper.GetString(utils.DhCoreAccessToken),
		},
	}

	ctx := context.Background()

	// Nuovo CrudService al posto del vecchio GetService
	crud, err := crudsvc.NewCrudService(ctx, cfg)
	if err != nil {
		return fmt.Errorf("sdk init failed: %w", err)
	}

	// Chiamata SDK (replica: se manca id usa name + versions=latest)
	body, _, err := crud.Get(ctx, crudsvc.GetRequest{
		ResourceRequest: crudsvc.ResourceRequest{
			Project:  project,
			Endpoint: endpoint,
		},
		ID:   id,
		Name: name,
	})
	if err != nil {
		return fmt.Errorf("error in request: %w", err)
	}

	switch format {
	case "short":
		return printShort(body)
	case "json":
		return printJson(id, body)
	case "yaml":
		utils.PrintCommentForYaml(env, resource, output, project, name, id)
		return printYaml(id, body)
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
}

func printShort(src []byte) error {
	var m map[string]interface{}
	if err := json.Unmarshal(src, &m); err != nil {
		return err
	}

	m, err := utils.GetFirstIfList(m)
	if err != nil {
		return err
	}

	fmt.Printf("%-12s %v\n", "Name:", m["name"])

	if status, ok := m["status"].(map[string]interface{}); ok {
		fmt.Printf("%-12s %v\n", "State:", status["state"])
	}

	fmt.Printf("%-12s %v\n", "Kind:", m["kind"])
	fmt.Printf("%-12s %v\n", "ID:", m["id"])
	fmt.Printf("%-12s %v\n", "Key:", m["key"])

	if meta, ok := m["metadata"].(map[string]interface{}); ok {
		fmt.Printf("%-12s %v\n", "Created on:", meta["created"])
		fmt.Printf("%-12s %v\n", "Created by:", meta["created_by"])
		fmt.Printf("%-12s %v\n", "Updated on:", meta["updated"])
		fmt.Printf("%-12s %v\n", "Updated by:", meta["updated_by"])
	}

	return nil
}

func printJson(id string, src []byte) error {
	var jsonData []byte = src
	if id == "" {
		var m map[string]interface{}
		if err := json.Unmarshal(src, &m); err != nil {
			return err
		}

		first, err := utils.GetFirstIfList(m)
		if err != nil {
			return err
		}

		out, err := json.Marshal(first)
		if err != nil {
			return err
		}
		jsonData = out
	}

	var pretty bytes.Buffer
	if err := json.Indent(&pretty, jsonData, "", "    "); err != nil {
		return err
	}
	fmt.Println(pretty.String())
	return nil
}

func printYaml(id string, src []byte) error {
	var yamlData []byte

	if id == "" {
		var m map[string]interface{}
		if err := json.Unmarshal(src, &m); err != nil {
			return err
		}

		first, err := utils.GetFirstIfList(m)
		if err != nil {
			return err
		}

		out, err := yaml.Marshal(first)
		if err != nil {
			return err
		}

		yamlData = out
	} else {
		out, err := yaml.JSONToYAML(src)
		if err != nil {
			return err
		}
		yamlData = out
	}

	fmt.Println(string(yamlData))
	return nil
}
