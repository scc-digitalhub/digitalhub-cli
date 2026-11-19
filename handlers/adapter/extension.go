// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"dhcli/handlers/utils"
	"dhcli/keys"

	"github.com/scc-digitalhub/digitalhub-cli-sdk/sdk/config"
	"github.com/spf13/viper"
	"sigs.k8s.io/yaml"
)

type ExtensionOptions struct {
	Action, Project, Parent, ParentID, ID, File, Output string
	ResetID                                             bool
}

type extension map[string]json.RawMessage

func ExtensionHandler(ctx context.Context, out io.Writer, opts ExtensionOptions) error {
	endpoint, err := extensionEndpoint(opts)
	if err != nil {
		return err
	}

	utils.CheckUpdateEnvironment()
	utils.CheckApiLevel(keys.ApiLevelKey, keys.MinApiLevel, 0)
	if err := utils.CheckCredentials(); err != nil {
		return err
	}
	core := config.CoreConfig{
		BaseURL: viper.GetString(keys.DhCoreEndpoint), APIVersion: viper.GetString(keys.DhCoreApiVersion),
		AccessToken: viper.GetString(keys.DhCoreAccessToken),
	}
	if core.AccessToken == "" {
		core.BasicAuthUsername = viper.GetString(keys.DhCoreUser)
		core.BasicAuthPassword = viper.GetString(keys.DhCorePassword)
	}
	if core.BaseURL == "" || core.APIVersion == "" {
		return fmt.Errorf("invalid core config: endpoint and API version are required")
	}
	httpCore := config.NewHTTPCore(utils.GetDebugHTTPClient(), core)
	return executeExtension(ctx, out, httpCore, strings.TrimRight(core.BaseURL, "/")+"/api/"+url.PathEscape(core.APIVersion)+endpoint, opts)
}

func extensionEndpoint(opts ExtensionOptions) (string, error) {
	var resource string
	for plural, aliases := range keys.Resources {
		if plural == opts.Parent {
			resource = plural
		}
		for _, alias := range aliases {
			if alias == opts.Parent {
				resource = plural
			}
		}
	}
	switch resource {
	case "artifacts", "dataitems", "functions", "models", "projects", "workflows", "runs":
	default:
		return "", fmt.Errorf("unsupported parent resource %q", opts.Parent)
	}
	parentID := opts.ParentID
	if parentID == "" {
		return "", fmt.Errorf("parent ID is required for %s extensions", opts.Parent)
	}
	if resource == "projects" {
		return "/projects/" + url.PathEscape(parentID) + "/extensions", nil
	}
	if opts.Project == "" {
		return "", fmt.Errorf("project is required for %s extensions", opts.Parent)
	}
	return "/-/" + url.PathEscape(opts.Project) + "/" + resource + "/" + url.PathEscape(parentID) + "/extensions", nil
}

func executeExtension(ctx context.Context, out io.Writer, core config.CoreHTTP, endpoint string, opts ExtensionOptions) error {
	var input extension
	var inputID string
	if opts.Action == "create" || opts.Action == "update" {
		data, err := os.ReadFile(opts.File)
		if err != nil {
			return fmt.Errorf("read extension file: %w", err)
		}
		data, err = yaml.YAMLToJSONStrict(data)
		if err != nil {
			return fmt.Errorf("parse extension YAML: %w", err)
		}
		if err := json.Unmarshal(data, &input); err != nil || input == nil {
			return fmt.Errorf("extension file must contain an object")
		}
		if _, err := extensionName(input); err != nil {
			return fmt.Errorf("extension file: %w", err)
		}
		if opts.Action == "create" {
			if opts.ResetID {
				delete(input, "id")
			} else if id, ok := input["id"]; ok && string(id) != "null" {
				if err := json.Unmarshal(id, &inputID); err != nil {
					return fmt.Errorf("extension file ID must be a string")
				}
			}
		}
	}
	if opts.Action == "get" || opts.Action == "list" {
		switch opts.Output {
		case "short", "json", "yaml", "yml":
		default:
			return fmt.Errorf("unknown output format %q", opts.Output)
		}
	}

	data, _, err := core.Do(ctx, "GET", endpoint, nil)
	if err != nil {
		return fmt.Errorf("fetch extensions: %w", err)
	}
	var extensions []extension
	if err := json.Unmarshal(data, &extensions); err != nil || extensions == nil {
		return fmt.Errorf("invalid extensions response: expected an array")
	}
	index := -1
	for i, ext := range extensions {
		id, err := extensionID(ext)
		if err != nil {
			return fmt.Errorf("invalid extension at index %d: %w", i, err)
		}
		if opts.Action == "create" && inputID != "" && inputID == id {
			return fmt.Errorf("extension ID %q already exists on this parent", inputID)
		}
		if id == opts.ID && opts.ID != "" {
			if index != -1 {
				return fmt.Errorf("multiple extensions with ID %q", opts.ID)
			}
			index = i
		}
	}

	switch opts.Action {
	case "list":
		return printExtensions(out, extensions, opts.Output)
	case "get":
		if index < 0 {
			return fmt.Errorf("extension %q not found", opts.ID)
		}
		return printExtensions(out, extensions[index], opts.Output)
	case "create":
		extensions = append(extensions, input)
	case "update":
		if index < 0 {
			return fmt.Errorf("extension %q not found", opts.ID)
		}
		if id, ok := extensions[index]["id"]; ok {
			if supplied, exists := input["id"]; exists {
				var suppliedID, existingID string
				if json.Unmarshal(supplied, &suppliedID) != nil || json.Unmarshal(id, &existingID) != nil || suppliedID != existingID {
					return fmt.Errorf("extension file ID does not match existing extension")
				}
			}
			input["id"] = id
		}
		extensions[index] = input
	case "delete":
		if index < 0 {
			return fmt.Errorf("extension %q not found", opts.ID)
		}
		extensions = append(extensions[:index], extensions[index+1:]...)
	default:
		return fmt.Errorf("unknown extension action %q", opts.Action)
	}
	data, err = json.Marshal(extensions)
	if err != nil {
		return fmt.Errorf("encode extensions: %w", err)
	}
	if _, _, err := core.Do(ctx, "PUT", endpoint, data); err != nil {
		return fmt.Errorf("store extensions: %w", err)
	}
	_, err = fmt.Fprintf(out, "Extension %s succeeded.\n", opts.Action)
	return err
}

func extensionName(ext extension) (string, error) {
	var name string
	if err := json.Unmarshal(ext["name"], &name); err != nil || strings.TrimSpace(name) == "" {
		return "", fmt.Errorf("name must be a non-empty string")
	}
	return name, nil
}

func extensionID(ext extension) (string, error) {
	var id string
	if err := json.Unmarshal(ext["id"], &id); err != nil || strings.TrimSpace(id) == "" {
		return "", fmt.Errorf("ID must be a non-empty string")
	}
	return id, nil
}

func printExtensions(out io.Writer, value any, format string) error {
	if format == "short" {
		switch v := value.(type) {
		case extension:
			name, err := extensionName(v)
			if err != nil {
				return err
			}
			_, err = fmt.Fprintln(out, name)
			return err
		case []extension:
			for _, ext := range v {
				name, err := extensionName(ext)
				if err != nil {
					return err
				}
				if _, err := fmt.Fprintln(out, name); err != nil {
					return err
				}
			}
			return nil
		}
	}
	data, err := json.MarshalIndent(value, "", "    ")
	if err != nil {
		return err
	}
	if format == "yaml" || format == "yml" {
		data, err = yaml.JSONToYAML(data)
		if err != nil {
			return err
		}
	}
	_, err = fmt.Fprintln(out, strings.TrimSuffix(string(data), "\n"))
	return err
}
