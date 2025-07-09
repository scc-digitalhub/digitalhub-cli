// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
	"sigs.k8s.io/yaml"

	"dhcli/utils"
)

func ListResourcesHandler(env string, output string, project string, name string, kind string, state string, resource string) error {
	endpoint := utils.TranslateEndpoint(resource)

	utils.CheckUpdateEnvironment()
	utils.CheckApiLevel(utils.ApiLevelKey, utils.ListMin, utils.ListMax)

	format := utils.TranslateFormat(output)

	if endpoint != "projects" && project == "" {
		return errors.New("project is mandatory when listing resources other than projects")
	}

	// Build query params
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

	// Fetch first page
	elements, _, err := fetchAllPages(project, endpoint, params)
	if err != nil {
		return fmt.Errorf("failed to fetch list: %w", err)
	}

	// Output
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

func fetchAllPages(project, endpoint string, params map[string]string) ([]interface{}, int, error) {
	var (
		elements   []interface{}
		currentPg  int
		totalPages int
	)

	for {
		url := utils.BuildCoreUrl(project, endpoint, "", params)
		req := utils.PrepareRequest("GET", url, nil, viper.GetString("access_token"))
		body, err := utils.DoRequest(req)
		if err != nil {
			return nil, 0, err
		}

		m := map[string]interface{}{}
		if err := json.Unmarshal(body, &m); err != nil {
			return nil, 0, fmt.Errorf("json parsing failed: %w", err)
		}

		pageList := m["content"].([]interface{})
		elements = append(elements, pageList...)

		pg := m["pageable"].(map[string]interface{})
		currentPg = int(reflect.ValueOf(pg["pageNumber"]).Float())
		totalPages = int(reflect.ValueOf(m["totalPages"]).Float())

		if currentPg >= totalPages-1 {
			break
		}
		params["page"] = strconv.Itoa(currentPg + 1)
	}

	return elements, totalPages, nil
}

func printShortList(resources []interface{}) {
	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{"NAME", "ID", "KIND", "UPDATED", "STATE", "LABELS"})

	for _, ri := range resources {
		m := ri.(map[string]interface{})
		name := m["name"].(string)
		id := m["id"].(string)
		kind := m["kind"].(string)

		updated := ""
		labels := ""
		if md, ok := m["metadata"].(map[string]interface{}); ok {
			if u, ok := md["updated"].(string); ok {
				updated = u
			}
			if lb, ok2 := md["labels"].([]interface{}); ok2 {
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

		table.Append([]string{name, id, kind, updated, state, labels})
	}

	table.Render()
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
