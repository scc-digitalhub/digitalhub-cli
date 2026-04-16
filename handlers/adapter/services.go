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
	"text/tabwriter"

	"github.com/scc-digitalhub/digitalhub-cli-sdk/sdk/config"

	crudsvc "github.com/scc-digitalhub/digitalhub-cli-sdk/sdk/services/crud"

	"github.com/spf13/viper"
	"sigs.k8s.io/yaml"

	"dhcli/handlers/utils"
)

// ListServicesHandler lists runs with action=serve filter
func ListServicesHandler(env string, output string, project string, name string, kind string, state string) error {
	endpoint := "runs"

	utils.CheckUpdateEnvironment()
	utils.CheckApiLevel(utils.ApiLevelKey, utils.ListMin, utils.ListMax)

	format := utils.TranslateFormat(output)

	if project == "" {
		return errors.New("project is mandatory when listing services")
	}

	// Config SDK (retrocompatibile: legge da viper/ini/env)
	cfg := config.Config{
		Core: config.CoreConfig{
			BaseURL:     viper.GetString(utils.DhCoreEndpoint),
			APIVersion:  viper.GetString(utils.DhCoreApiVersion),
			AccessToken: viper.GetString(utils.DhCoreAccessToken),
		},
		HTTPClient: utils.GetDebugHTTPClient(),
	}

	ctx := context.Background()

	// Nuovo CrudService
	crud, err := crudsvc.NewCrudService(ctx, cfg)
	if err != nil {
		return fmt.Errorf("sdk init failed: %w", err)
	}

	// Query params with action=serve filter
	// Default state to RUNNING if not specified
	if state == "" {
		state = "RUNNING"
	}
	params := map[string]string{
		"name":   name,
		"kind":   kind,
		"state":  state,
		"action": "serve",
		"size":   "200",
		"sort":   "updated,asc",
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
		return fmt.Errorf("failed to fetch services list: %w", err)
	}

	// Output IDENTICO
	switch format {
	case "short":
		printShortServicesList(elements)
	case "json":
		printJSONServicesList(elements)
	case "yaml":
		utils.PrintCommentForYaml(env, "runs", output, project, name, kind, state)
		printYAMLServicesList(elements)
	default:
		return fmt.Errorf("unknown format: %s", format)
	}

	return nil
}

// parseFunctionName extracts the name from a function URI
// Format: <kind>://<project>/<name>:<version>
// Returns only the <name> part
func parseFunctionName(functionURI string) string {
	if functionURI == "" {
		return ""
	}

	// Find the last occurrence of '/'
	lastSlashIdx := strings.LastIndex(functionURI, "/")
	if lastSlashIdx == -1 {
		return functionURI
	}

	// Get everything after the last '/'
	nameWithVersion := functionURI[lastSlashIdx+1:]

	// Find the first occurrence of ':' to separate name from version
	colonIdx := strings.Index(nameWithVersion, ":")
	if colonIdx == -1 {
		return nameWithVersion
	}

	return nameWithVersion[:colonIdx]
}

func printShortServicesList(resources []interface{}) {
	// Create tabwriter with proper spacing
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	// Print header
	fmt.Fprintln(w, "NAME\tID\tFUNCTION\tKIND\tSERVICE\tUPDATED\tSTATE")

	for _, ri := range resources {
		m := ri.(map[string]interface{})
		name := fmt.Sprintf("%v", m["name"])
		id := fmt.Sprintf("%v", m["id"])
		kind := fmt.Sprintf("%v", m["kind"])

		specFunction := ""
		if spec, ok := m["spec"].(map[string]interface{}); ok {
			if fn, ok := spec["function"].(string); ok {
				specFunction = parseFunctionName(fn)
			}
		}

		serviceURL := ""
		if status, ok := m["status"].(map[string]interface{}); ok {
			if svc, ok := status["service"].(map[string]interface{}); ok {
				if url, ok := svc["url"].(string); ok {
					serviceURL = url
				}
			}
		}

		updated := ""
		if md, ok := m["metadata"].(map[string]interface{}); ok {
			if u, ok := md["updated"].(string); ok {
				updated = u
			}
		}

		state := ""
		if st, ok := m["status"].(map[string]interface{}); ok {
			if s, ok := st["state"].(string); ok {
				state = s
			}
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n", name, id, specFunction, kind, serviceURL, updated, state)
	}

	w.Flush()
}

func printJSONServicesList(resources []interface{}) {
	out, err := json.MarshalIndent(resources, "", "    ")
	if err != nil {
		log.Printf("Error serializing JSON: %v", err)
		os.Exit(1)
	}
	fmt.Println(string(out))
}

func printYAMLServicesList(resources []interface{}) {
	out, err := yaml.Marshal(resources)
	if err != nil {
		log.Printf("Error serializing YAML: %v", err)
		os.Exit(1)
	}
	fmt.Println(string(out))
}
