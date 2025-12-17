// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/scc-digitalhub/digitalhub-cli-sdk/sdk/config"

	crudsvc "github.com/scc-digitalhub/digitalhub-cli-sdk/sdk/services/crud"

	"github.com/spf13/viper"
	"sigs.k8s.io/yaml"

	"github.com/scc-digitalhub/digitalhub-cli-sdk/sdk/utils"
)

func ListResourcesHandler(env string, output string, project string, name string, kind string, state string, resource string) error {
	endpoint := utils.TranslateEndpoint(resource)

	utils.CheckUpdateEnvironment()
	utils.CheckApiLevel(utils.ApiLevelKey, utils.ListMin, utils.ListMax)

	format := utils.TranslateFormat(output)

	if endpoint != "projects" && project == "" {
		return errors.New("project is mandatory when performing this operation on resources other than projects")
	}

	// Config SDK (retrocompatibile: legge da viper/ini/env)
	cfg := config.Config{
		Core: config.CoreConfig{
			BaseURL:     viper.GetString(utils.DhCoreEndpoint),
			APIVersion:  viper.GetString(utils.DhCoreApiVersion),
			AccessToken: viper.GetString(utils.DhCoreAccessToken),
		},
	}

	ctx := context.Background()

	// Nuovo CrudService
	crud, err := crudsvc.NewCrudService(ctx, cfg)
	if err != nil {
		return fmt.Errorf("sdk init failed: %w", err)
	}

	// Query params => identici all’originale
	params := map[string]string{
		"name":  name,
		"kind":  kind,
		"state": state,
		"size":  "200",
		"sort":  "updated,asc",
	}
	if name != "" {
		params["versions"] = "all"
	}

	// Paging identico: FetchAllPages diventa ListAllPages
	elements, _, err := crud.ListAllPages(ctx, crudsvc.ListRequest{
		ResourceRequest: crudsvc.ResourceRequest{
			Project:  project,
			Resource: endpoint,
		},
		Params: params,
	})
	if err != nil {
		return fmt.Errorf("failed to fetch list: %w", err)
	}

	// Output IDENTICO
	switch format {
	case "short":
		printShortList(elements)
	case "json":
		printJSONList(elements)
	case "yaml":
		utils.PrintCommentForYaml(env, resource, output, project, name, kind, state)
		printYAMLList(elements)
	default:
		return fmt.Errorf("unknown format: %s", format)
	}

	return nil
}

func printShortList(resources []interface{}) {
	printShortLineList("NAME", "ID", "KIND", "UPDATED", "STATE", "LABELS")

	for _, ri := range resources {
		m := ri.(map[string]interface{})
		name := fmt.Sprintf("%v", m["name"])
		id := fmt.Sprintf("%v", m["id"])
		kind := fmt.Sprintf("%v", m["kind"])

		updated := ""
		labels := ""
		if md, ok := m["metadata"].(map[string]interface{}); ok {
			if u, ok := md["updated"].(string); ok {
				updated = u
			}
			if lb, ok := md["labels"].([]interface{}); ok {
				strs := []string{}
				for _, v := range lb {
					strs = append(strs, fmt.Sprint(v))
				}
				labels = strings.Join(strs, ", ")
			}
		}

		state := ""
		if st, ok := m["status"].(map[string]interface{}); ok {
			if s, ok := st["state"].(string); ok {
				state = s
			}
		}

		printShortLineList(name, id, kind, updated, state, labels)
	}
}

func printShortLineList(rName string, rId string, rKind string, rUpdated string, rState string, rLabels string) {
	fmt.Printf("%-36s%-36s%-24s%-30s%-12s%s\n", rName, rId, rKind, rUpdated, rState, rLabels)
}

func printJSONList(resources []interface{}) {
	out, err := json.MarshalIndent(resources, "", "    ")
	if err != nil {
		log.Printf("Error serializing JSON: %v", err)
		os.Exit(1)
	}
	fmt.Println(string(out))
}

func printYAMLList(resources []interface{}) {
	out, err := yaml.Marshal(resources)
	if err != nil {
		log.Printf("Error serializing YAML: %v", err)
		os.Exit(1)
	}
	fmt.Println(string(out))
}
