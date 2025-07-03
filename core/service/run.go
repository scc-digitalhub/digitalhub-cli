// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"dhcli/utils"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"

	"gopkg.in/ini.v1"
	"sigs.k8s.io/yaml"
)

func RunHandler(env string, project string, functionName string, functionId string, filePath string, task string) error {
	endpoint := utils.TranslateEndpoint("run")

	// Load environment and check API level requirements
	cfg, section := utils.LoadIniConfig([]string{env})
	utils.CheckUpdateEnvironment(cfg, section)
	utils.CheckApiLevel(section, utils.CreateMin, utils.CreateMax)

	if project == "" {
		return errors.New("Project not specified.")
	}

	// Get function kind and key
	functionKind, functionKey, err := getFunctionKey(section, project, functionId, functionName)
	log.Printf("FUNCTION: %v\n", functionKey)
	if err != nil {
		return err
	}

	// Get or create task
	taskKey, err := getTaskKey(section, project, functionKey, task)
	log.Printf("TASK: %v\n", taskKey)
	if err != nil {
		var err error
		taskKey, err = createTask(section, project, functionKey, task)
		if err != nil {
			return err
		}
	}

	// Run specification
	jsonMap := map[string]interface{}{}
	var spec map[string]interface{}

	if filePath != "" {
		// Read file
		file, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}

		// Convert YAML to JSON
		jsonBytes, err := yaml.YAMLToJSON(file)

		// Convert to map
		var inputJson map[string]interface{}
		err = json.Unmarshal(jsonBytes, &inputJson)
		if err != nil {
			return err
		}
		spec = inputJson["spec"].(map[string]interface{})
	} else {
		spec = map[string]interface{}{}
	}

	jsonMap["kind"] = fmt.Sprintf("%v+run", functionKind)
	jsonMap["project"] = project

	spec["task"] = taskKey
	spec["function"] = functionKey
	jsonMap["spec"] = spec

	// Marshal
	jsonBody, err := json.Marshal(jsonMap)
	if err != nil {
		log.Printf("Failed to marshal: %v\n", err)
		os.Exit(1)
	}

	// Request
	method := "POST"
	url := utils.BuildCoreUrl(section, project, endpoint, "", nil)
	req := utils.PrepareRequest(method, url, jsonBody, section.Key("access_token").String())
	_, err = utils.DoRequest(req)
	if err != nil {
		return err
	}

	log.Println("Created successfully.")
	return nil
}

func getFunctionKey(section *ini.Section, project string, id string, name string) (string, string, error) {
	var function map[string]interface{}
	if id != "" {
		// Get function by ID
		method := "GET"
		url := utils.BuildCoreUrl(section, project, "functions", id, nil)
		req := utils.PrepareRequest(method, url, nil, section.Key("access_token").String())
		resp, err := utils.DoRequest(req)
		if err != nil {
			return "", "", err
		}
		if err := json.Unmarshal(resp, &function); err != nil {
			return "", "", err
		}
	} else if name != "" {
		// Get latest function by name
		method := "GET"
		url := utils.BuildCoreUrl(section, project, "functions", "", map[string]string{"name": name})
		req := utils.PrepareRequest(method, url, nil, section.Key("access_token").String())
		resp, err := utils.DoRequest(req)
		if err != nil {
			return "", "", err
		}
		var m map[string]interface{}
		if err := json.Unmarshal(resp, &m); err != nil {
			return "", "", err
		}
		latest, err := utils.GetFirstIfList(m)
		if err != nil {
			return "", "", err
		}
		function = latest
	} else {
		return "", "", errors.New("You must provide the name or ID of the function to run.")
	}

	if kind, ok := function["kind"]; ok {
		if id, ok := function["id"]; ok {
			if fname, ok := function["name"]; ok {
				return kind.(string), fmt.Sprintf("%v://%v/%v:%v", kind, project, fname, id), nil
			}
		}
	}

	return "", "", errors.New("Unable to obtain function key.")
}

func getTaskKey(section *ini.Section, project string, functionKey string, task string) (string, error) {
	// Perform request
	method := "GET"
	params := map[string]string{"function": functionKey}
	url := utils.BuildCoreUrl(section, project, "tasks", "", params)
	req := utils.PrepareRequest(method, url, nil, section.Key("access_token").String())

	resp, err := utils.DoRequest(req)
	if err != nil {
		return "", err
	}

	var m map[string]interface{}
	if err := json.Unmarshal(resp, &m); err != nil {
		return "", err
	}

	if content, ok := m["content"]; ok && reflect.ValueOf(content).Kind() == reflect.Slice {
		contentSlice := content.([]interface{})
		for _, taskItem := range contentSlice {
			taskMap := taskItem.(map[string]interface{})
			if kind, ok := taskMap["kind"]; ok && kind.(string) == task {
				if id, ok := taskMap["id"]; ok {
					return fmt.Sprintf("%v://%v/%v", kind, project, id), nil
				}
			}
		}
	}
	return "", errors.New("Unable to obtain task key.")
}

func createTask(section *ini.Section, project string, function string, task string) (string, error) {
	method := "POST"
	url := utils.BuildCoreUrl(section, project, "tasks", "", nil)

	// Body
	reqBody := map[string]interface{}{}
	reqBody["kind"] = task
	reqBody["project"] = project

	spec := map[string]interface{}{}
	spec["function"] = function
	reqBody["spec"] = spec

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", nil
	}

	// Perform request
	req := utils.PrepareRequest(method, url, jsonBody, section.Key("access_token").String())
	resp, err := utils.DoRequest(req)
	if err != nil {
		return "", err
	}

	var m map[string]interface{}
	if err := json.Unmarshal(resp, &m); err != nil {
		return "", err
	}

	latest, err := utils.GetFirstIfList(m)
	if err != nil {
		return "", err
	}

	if kind, ok := latest["kind"]; ok {
		if id, ok := latest["id"]; ok {
			return fmt.Sprintf("%v://%v/%v", kind, project, id), nil
		}
	}

	return "", errors.New("Unable to obtain task key.")
}
