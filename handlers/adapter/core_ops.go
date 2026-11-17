// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	"dhcli/handlers/utils"
	"dhcli/keys"

	"github.com/scc-digitalhub/digitalhub-cli-sdk/sdk/config"
	"github.com/spf13/viper"
	"sigs.k8s.io/yaml"
)

// CreateOptions configures resource creation across CLI and MCP callers.
type CreateOptions struct {
	Env            string
	Output         string
	Project        string
	Name           string
	FilePath       string
	Manifest       any
	SourceFile     string
	SecretValue    string
	SecretValueSet bool
	Task           string
	TargetName     string
	TargetID       string
	Schedule       string
	TriggerKind    string
	ResetID        bool
	Resource       string
}

// UpdateOptions configures resource updates across CLI and MCP callers.
type UpdateOptions struct {
	Env            string
	Output         string
	Project        string
	Resource       string
	ID             string
	FilePath       string
	Manifest       any
	SourceFile     string
	SecretValue    string
	SecretValueSet bool
}

// RunOptions configures function and workflow execution across CLI and MCP callers.
type RunOptions struct {
	Env        string
	Output     string
	Project    string
	Task       string
	Name       string
	ID         string
	FilePath   string
	Spec       map[string]any
	IsWorkflow bool
}

// NewCoreConfig builds the SDK config from Viper.
func NewCoreConfig() config.Config {
	token, user, password := utils.CurrentHTTPAuth()
	return config.Config{
		Core: config.CoreConfig{
			BaseURL:           strings.TrimRight(viper.GetString(keys.DhCoreEndpoint), "/"),
			APIVersion:        viper.GetString(keys.DhCoreApiVersion),
			AccessToken:       token,
			BasicAuthUsername: user,
			BasicAuthPassword: password,
		},
		HTTPClient: utils.GetDebugHTTPClient(),
	}
}

// NewCoreHTTP returns a CoreHTTP client configured from Viper.
func NewCoreHTTP() config.CoreHTTP {
	cfg := NewCoreConfig()
	return encodedCoreHTTP{CoreHTTP: config.NewHTTPCore(cfg.HTTPClient, cfg.Core)}
}

type encodedCoreHTTP struct{ config.CoreHTTP }

func (c encodedCoreHTTP) BuildURL(project, resource, id string, params map[string]string) string {
	base := c.CoreHTTP.BuildURL(project, resource, id, nil)
	values := url.Values{}
	for key, value := range params {
		if value != "" {
			values.Set(key, value)
		}
	}
	if query := values.Encode(); query != "" {
		base += "?" + query
	}
	return base
}

func prepareEnvironment(ctx context.Context, env string, minLevel, maxLevel int) error {
	if env != "" {
		if err := utils.ActivateEnvironment(env); err != nil {
			return err
		}
	}
	if err := utils.CheckUpdateEnvironmentContext(ctx); err != nil {
		return err
	}
	if err := utils.CheckApiLevelErr(keys.ApiLevelKey, minLevel, maxLevel); err != nil {
		return err
	}
	return utils.CheckCredentialsContext(ctx)
}

// ParseManifestInput parses a resource manifest from a file path or inline object/string.
func ParseManifestInput(filePath string, inline any) (map[string]any, error) {
	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read YAML file: %w", err)
		}
		return unmarshalYAMLOrJSONMap(data)
	}
	if inline == nil {
		return nil, nil
	}
	switch v := inline.(type) {
	case map[string]any:
		// Deep copy via JSON round-trip to avoid mutating caller map
		b, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal inline manifest: %w", err)
		}
		var out map[string]any
		if err := json.Unmarshal(b, &out); err != nil {
			return nil, fmt.Errorf("failed to parse inline manifest: %w", err)
		}
		return out, nil
	case string:
		trimmed := strings.TrimSpace(v)
		if trimmed == "" {
			return nil, nil
		}
		return unmarshalYAMLOrJSONMap([]byte(trimmed))
	case []byte:
		if len(bytes.TrimSpace(v)) == 0 {
			return nil, nil
		}
		return unmarshalYAMLOrJSONMap(v)
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("unsupported manifest type %T: %w", inline, err)
		}
		var out map[string]any
		if err := json.Unmarshal(b, &out); err != nil {
			return nil, fmt.Errorf("failed to parse manifest: %w", err)
		}
		return out, nil
	}
}

func unmarshalYAMLOrJSONMap(data []byte) (map[string]any, error) {
	jsonBytes, err := yaml.YAMLToJSON(data)
	if err != nil {
		return nil, fmt.Errorf("yaml to json failed: %w", err)
	}
	var m map[string]any
	if err := json.Unmarshal(jsonBytes, &m); err != nil {
		return nil, fmt.Errorf("failed to parse after JSON conversion: %w", err)
	}
	return m, nil
}

// InjectSourceFile reads sourceFile from disk, Base64-encodes it, and sets spec.source.base64.
func InjectSourceFile(manifest map[string]any, sourceFile string) error {
	if sourceFile == "" {
		return nil
	}
	raw, err := os.ReadFile(sourceFile)
	if err != nil {
		return fmt.Errorf("failed to read source file %q: %w", sourceFile, err)
	}
	encoded := base64.StdEncoding.EncodeToString(raw)

	spec, ok := manifest["spec"].(map[string]any)
	if !ok || spec == nil {
		spec = map[string]any{}
		manifest["spec"] = spec
	}
	src, ok := spec["source"].(map[string]any)
	if !ok || src == nil {
		src = map[string]any{}
		spec["source"] = src
	}
	src["base64"] = encoded

	if _, hasLang := src["lang"]; !hasLang {
		lower := strings.ToLower(sourceFile)
		switch {
		case strings.HasSuffix(lower, ".py"):
			src["lang"] = "python"
		case strings.HasSuffix(lower, ".sql"):
			src["lang"] = "sql"
		}
	}
	return nil
}

// CreateResource executes resource creation (including smart secret and trigger handling) and returns the raw JSON response.
func CreateResource(ctx context.Context, opts CreateOptions) ([]byte, error) {
	endpoint, err := utils.ResolveEndpoint(opts.Resource)
	if err != nil {
		return nil, err
	}

	if err := prepareEnvironment(ctx, opts.Env, keys.CreateMin, keys.CreateMax); err != nil {
		return nil, err
	}

	if endpoint != "projects" && opts.Project == "" {
		return nil, errors.New("project is mandatory when performing this operation on resources other than projects")
	}

	jsonMap, err := ParseManifestInput(opts.FilePath, opts.Manifest)
	if err != nil {
		return nil, err
	}

	coreHTTP := NewCoreHTTP()

	switch endpoint {
	case "projects":
		if jsonMap == nil {
			if opts.Name == "" {
				return nil, errors.New("must provide either an input file/manifest or a name when creating a project")
			}
			jsonMap = map[string]any{
				"name": opts.Name,
			}
		} else if opts.Name != "" && jsonMap["name"] == nil {
			jsonMap["name"] = opts.Name
		}

	case "secrets":
		return createSecretResource(ctx, coreHTTP, opts, jsonMap)

	case "triggers":
		var err error
		jsonMap, err = prepareTriggerManifest(ctx, coreHTTP, opts, jsonMap)
		if err != nil {
			return nil, err
		}

	default:
		if jsonMap == nil {
			return nil, errors.New("input file or manifest not specified")
		}
		if opts.Name != "" && jsonMap["name"] == nil {
			jsonMap["name"] = opts.Name
		}
	}

	if err := InjectSourceFile(jsonMap, opts.SourceFile); err != nil {
		return nil, err
	}

	delete(jsonMap, "user")
	if endpoint != "projects" {
		jsonMap["project"] = opts.Project
	}
	if opts.ResetID {
		delete(jsonMap, "id")
	}

	body, err := json.Marshal(jsonMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal: %w", err)
	}

	url := coreHTTP.BuildURL(opts.Project, endpoint, "", nil)
	respBody, _, err := coreHTTP.Do(ctx, "POST", url, body)
	if err != nil {
		return nil, err
	}
	return respBody, nil
}

func createSecretResource(ctx context.Context, coreHTTP config.CoreHTTP, opts CreateOptions, jsonMap map[string]any) ([]byte, error) {
	secretVal := opts.SecretValue
	secretSet := opts.SecretValueSet || secretVal != ""

	if jsonMap == nil {
		if opts.Name == "" {
			return nil, errors.New("must provide either --name (-n) or an input file/manifest when creating a secret")
		}
		jsonMap = map[string]any{
			"name":    opts.Name,
			"kind":    "secret",
			"project": opts.Project,
			"spec": map[string]any{
				"path":     "secret://" + opts.Name,
				"provider": "kubernetes",
			},
		}
	} else {
		if opts.Name != "" && jsonMap["name"] == nil {
			jsonMap["name"] = opts.Name
		}
		if jsonMap["kind"] == nil || jsonMap["kind"] == "" {
			jsonMap["kind"] = "secret"
		}
		spec, ok := jsonMap["spec"].(map[string]any)
		if !ok || spec == nil {
			spec = map[string]any{}
			jsonMap["spec"] = spec
		}
		var err error
		secretVal, secretSet, err = extractSecretValue(jsonMap, secretVal, secretSet)
		if err != nil {
			return nil, err
		}

		secretName, _ := jsonMap["name"].(string)
		if secretName == "" {
			return nil, errors.New("secret name is required")
		}
		if spec["path"] == nil || spec["path"] == "" {
			spec["path"] = "secret://" + secretName
		}
		if spec["provider"] == nil || spec["provider"] == "" {
			spec["provider"] = "kubernetes"
		}
	}

	delete(jsonMap, "user")
	jsonMap["project"] = opts.Project
	if opts.ResetID {
		delete(jsonMap, "id")
	}

	secretName, _ := jsonMap["name"].(string)
	if secretName == "" {
		return nil, errors.New("secret name is required")
	}

	body, err := json.Marshal(jsonMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal secret: %w", err)
	}

	url := coreHTTP.BuildURL(opts.Project, "secrets", "", nil)
	respBody, _, err := coreHTTP.Do(ctx, "POST", url, body)
	if err != nil {
		return nil, err
	}

	if secretSet {
		if err := PutSecretData(ctx, coreHTTP, opts.Project, secretName, secretVal); err != nil {
			return respBody, fmt.Errorf("secret entity created, but setting secret data failed: %w", err)
		}
	}

	return respBody, nil
}

func extractSecretValue(manifest map[string]any, value string, set bool) (string, bool, error) {
	maps := []map[string]any{manifest}
	if spec, ok := manifest["spec"].(map[string]any); ok {
		maps = append(maps, spec)
	}
	for _, m := range maps {
		for _, key := range []string{"secret_value", "value"} {
			if raw, exists := m[key]; exists {
				text, ok := raw.(string)
				if !ok {
					return "", false, fmt.Errorf("secret %s must be a string", key)
				}
				if !set {
					value, set = text, true
				}
				delete(m, key)
			}
		}
	}
	return value, set, nil
}

// PutSecretData writes the key/value pair to PUT /api/<ver>/-/<project>/secrets/data.
func PutSecretData(ctx context.Context, coreHTTP config.CoreHTTP, project, secretName, secretValue string) error {
	payload := map[string]string{
		secretName: secretValue,
	}
	dataBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	url := coreHTTP.BuildURL(project, "secrets/data", "", nil)
	_, _, err = coreHTTP.Do(ctx, "PUT", url, dataBytes)
	return err
}

func prepareTriggerManifest(ctx context.Context, coreHTTP config.CoreHTTP, opts CreateOptions, jsonMap map[string]any) (map[string]any, error) {
	if jsonMap == nil {
		if opts.Name == "" {
			return nil, errors.New("must provide either an input file/manifest or --name (-n) when creating a trigger")
		}
		jsonMap = map[string]any{
			"name":    opts.Name,
			"project": opts.Project,
		}
	} else if opts.Name != "" && jsonMap["name"] == nil {
		jsonMap["name"] = opts.Name
	}

	spec, ok := jsonMap["spec"].(map[string]any)
	if !ok || spec == nil {
		spec = map[string]any{}
		jsonMap["spec"] = spec
	}

	if opts.Schedule != "" && spec["schedule"] == nil {
		spec["schedule"] = opts.Schedule
	}

	if jsonMap["kind"] == nil || jsonMap["kind"] == "" {
		switch {
		case opts.TriggerKind != "":
			jsonMap["kind"] = opts.TriggerKind
		case spec["schedule"] != nil && spec["schedule"] != "":
			jsonMap["kind"] = "scheduler"
		case spec["key"] != nil && spec["key"] != "":
			jsonMap["kind"] = "lifecycle"
		default:
			jsonMap["kind"] = "scheduler"
		}
	}

	// Determine task kind or task URI
	taskRef, _ := spec["task"].(string)
	if opts.Task != "" {
		taskRef = opts.Task
	}
	if taskRef == "" {
		return nil, errors.New("trigger requires a task (set spec.task in YAML or pass --task <kind+action>)")
	}

	// Determine target executable (function or workflow)
	fnRef, _ := spec["function"].(string)
	wfRef, _ := spec["workflow"].(string)
	targetName := opts.TargetName
	targetID := opts.TargetID

	isWF := isWorkflowTask(taskRef) || wfRef != ""
	execField := "function"
	execResource := "functions"
	execRef := fnRef
	if isWF {
		execField = "workflow"
		execResource = "workflows"
		execRef = wfRef
	}

	if targetName == "" && targetID == "" && execRef != "" && !strings.Contains(execRef, "://") {
		targetName = execRef
		execRef = ""
	}

	if execRef == "" {
		if targetName == "" && targetID == "" {
			return nil, fmt.Errorf("trigger requires a target %s (set spec.%s in YAML or pass --target-name / --fn-name / --wf-name)", execField, execField)
		}
		_, resolvedKey, err := resolveExecutable(ctx, coreHTTP, opts.Project, execResource, targetID, targetName)
		if err != nil {
			return nil, err
		}
		execRef = resolvedKey
	}

	taskKey := taskRef
	if !strings.Contains(taskKey, "://") {
		resolvedTaskKey, err := getOrCreateTask(ctx, coreHTTP, opts.Project, execField, execRef, taskRef)
		if err != nil {
			return nil, err
		}
		taskKey = resolvedTaskKey
	}

	spec["task"] = taskKey
	spec[execField] = execRef

	tmpl, ok := spec["template"].(map[string]any)
	if !ok || tmpl == nil {
		tmpl = map[string]any{}
		spec["template"] = tmpl
	}
	tmpl["task"] = taskKey
	tmpl[execField] = execRef

	return jsonMap, nil
}

// UpdateResource updates a resource by ID (and optionally updates secret data) and returns the raw JSON response.
func UpdateResource(ctx context.Context, opts UpdateOptions) ([]byte, error) {
	endpoint, err := utils.ResolveEndpoint(opts.Resource)
	if err != nil {
		return nil, err
	}

	if err := prepareEnvironment(ctx, opts.Env, keys.UpdateMin, keys.UpdateMax); err != nil {
		return nil, err
	}

	if endpoint != "projects" && opts.Project == "" {
		return nil, errors.New("project is mandatory when performing this operation on resources other than projects")
	}
	if opts.ID == "" {
		return nil, errors.New("id is required")
	}

	jsonMap, err := ParseManifestInput(opts.FilePath, opts.Manifest)
	if err != nil {
		return nil, err
	}

	coreHTTP := NewCoreHTTP()
	secretVal := opts.SecretValue
	secretSet := opts.SecretValueSet || secretVal != ""
	if endpoint == "secrets" && jsonMap != nil {
		secretVal, secretSet, err = extractSecretValue(jsonMap, secretVal, secretSet)
		if err != nil {
			return nil, err
		}
	}

	// Allow updating only a secret's value if --value / secret_value is provided without a file
	if endpoint == "secrets" && jsonMap == nil && secretSet {
		getURL := coreHTTP.BuildURL(opts.Project, "secrets", opts.ID, nil)
		b, _, err := coreHTTP.Do(ctx, "GET", getURL, nil)
		if err != nil {
			return nil, err
		}
		var secMap map[string]any
		if err := json.Unmarshal(b, &secMap); err != nil {
			return nil, err
		}
		secName, _ := secMap["name"].(string)
		if secName == "" {
			return nil, errors.New("unable to determine secret name for secret value update")
		}
		if err := PutSecretData(ctx, coreHTTP, opts.Project, secName, secretVal); err != nil {
			return nil, err
		}
		return b, nil
	}

	if jsonMap == nil {
		return nil, errors.New("input file or manifest not specified")
	}

	if jsonMap["id"] != nil && fmt.Sprint(jsonMap["id"]) != opts.ID {
		return nil, fmt.Errorf("specified ID (%v) and ID found in manifest (%v) do not match", opts.ID, jsonMap["id"])
	}

	secretName := ""
	if endpoint == "secrets" && secretSet {
		secretName, _ = jsonMap["name"].(string)
		if secretName == "" {
			getURL := coreHTTP.BuildURL(opts.Project, "secrets", opts.ID, nil)
			data, _, err := coreHTTP.Do(ctx, "GET", getURL, nil)
			if err != nil {
				return nil, fmt.Errorf("resolve secret name before update: %w", err)
			}
			var existing struct {
				Name string `json:"name"`
			}
			if err := json.Unmarshal(data, &existing); err != nil {
				return nil, fmt.Errorf("resolve secret name before update: %w", err)
			}
			secretName = existing.Name
		}
		if secretName == "" {
			return nil, errors.New("unable to determine secret name for secret value update")
		}
	}

	if err := InjectSourceFile(jsonMap, opts.SourceFile); err != nil {
		return nil, err
	}

	delete(jsonMap, "user")
	if endpoint != "projects" {
		jsonMap["project"] = opts.Project
	}

	body, err := json.Marshal(jsonMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal: %w", err)
	}

	url := coreHTTP.BuildURL(opts.Project, endpoint, opts.ID, nil)
	respBody, status, err := coreHTTP.Do(ctx, "PUT", url, body)
	if err != nil {
		return nil, fmt.Errorf("update failed (status %d): %w", status, err)
	}

	if endpoint == "secrets" && secretSet {
		if err := PutSecretData(ctx, coreHTTP, opts.Project, secretName, secretVal); err != nil {
			return respBody, fmt.Errorf("secret entity updated, but setting secret data failed: %w", err)
		}
	}

	return respBody, nil
}

// isWorkflowTask returns true if taskKind belongs to a workflow runtime (hera or kfp).
func isWorkflowTask(taskKind string) bool {
	trimmed := strings.TrimSpace(taskKind)
	prefix, _, _ := strings.Cut(trimmed, "+")
	return prefix == "hera" || prefix == "kfp"
}

func taskToRunKind(task string) string {
	task = strings.TrimSpace(task)
	if task == "" {
		return task
	}
	if i := strings.IndexByte(task, ':'); i >= 0 {
		return task[:i] + ":run"
	}
	return task + ":run"
}

// ExecuteRun resolves the target function or workflow, ensures the task exists, creates the run, and returns the Run JSON.
func ExecuteRun(ctx context.Context, opts RunOptions) ([]byte, error) {
	if err := prepareEnvironment(ctx, opts.Env, keys.CreateMin, keys.CreateMax); err != nil {
		return nil, err
	}

	if opts.Project == "" {
		return nil, errors.New("project not specified")
	}
	if strings.TrimSpace(opts.Task) == "" {
		return nil, errors.New("task kind not specified")
	}

	inputSpec := map[string]any{}
	if opts.FilePath != "" || opts.Spec != nil {
		parsed, err := ParseManifestInput(opts.FilePath, opts.Spec)
		if err != nil {
			return nil, err
		}
		if parsed != nil {
			if s, ok := parsed["spec"].(map[string]any); ok {
				inputSpec = s
			} else {
				inputSpec = parsed
			}
		}
	}

	coreHTTP := NewCoreHTTP()
	origTaskKind := strings.TrimSpace(opts.Task)
	runKind := taskToRunKind(origTaskKind)

	isWF := opts.IsWorkflow || isWorkflowTask(origTaskKind)
	execResource := "functions"
	execField := "function"
	if isWF {
		execResource = "workflows"
		execField = "workflow"
	}

	_, execKey, err := resolveExecutable(ctx, coreHTTP, opts.Project, execResource, opts.ID, opts.Name)
	if err != nil {
		return nil, err
	}

	taskKey, err := getOrCreateTask(ctx, coreHTTP, opts.Project, execField, execKey, origTaskKind)
	if err != nil {
		return nil, err
	}

	spec := map[string]any{}
	for k, v := range inputSpec {
		spec[k] = v
	}
	spec["task"] = taskKey
	spec[execField] = execKey
	spec["local_execution"] = false

	body := map[string]any{
		"kind":    runKind,
		"project": opts.Project,
		"spec":    spec,
	}
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	url := coreHTTP.BuildURL(opts.Project, "runs", "", nil)
	respBody, status, err := coreHTTP.Do(ctx, "POST", url, data)
	if err != nil {
		return nil, fmt.Errorf("run creation failed (status %d): %w", status, err)
	}
	return respBody, nil
}

func resolveExecutable(ctx context.Context, coreHTTP config.CoreHTTP, project, resource, id, name string) (string, string, error) {
	singular := strings.TrimSuffix(resource, "s")
	var entity map[string]any

	if id != "" {
		url := coreHTTP.BuildURL(project, resource, id, nil)
		b, status, err := coreHTTP.Do(ctx, "GET", url, nil)
		if err != nil {
			return "", "", fmt.Errorf("get %s by id failed (status %d): %w", singular, status, err)
		}
		if err := json.Unmarshal(b, &entity); err != nil {
			return "", "", err
		}
	} else if name != "" {
		url := coreHTTP.BuildURL(project, resource, "", map[string]string{"name": name})
		b, status, err := coreHTTP.Do(ctx, "GET", url, nil)
		if err != nil {
			return "", "", fmt.Errorf("get %s by name failed (status %d): %w", singular, status, err)
		}
		var m map[string]any
		if err := json.Unmarshal(b, &m); err != nil {
			return "", "", err
		}
		first, err := utils.GetFirstIfList(m)
		if err != nil {
			return "", "", fmt.Errorf("%s %q not found in project %q", singular, name, project)
		}
		entity = first
	} else {
		return "", "", fmt.Errorf("you must provide the name or ID of the %s to run", singular)
	}

	kind, ok1 := entity["kind"].(string)
	idVal, ok2 := entity["id"]
	nameVal, ok3 := entity["name"]
	if !ok1 || !ok2 || !ok3 {
		return "", "", fmt.Errorf("unable to obtain %s key", singular)
	}
	return kind, fmt.Sprintf("%s://%s/%s:%s", kind, project, fmt.Sprint(nameVal), fmt.Sprint(idVal)), nil
}

func getOrCreateTask(ctx context.Context, coreHTTP config.CoreHTTP, project, execField, execKey, taskKind string) (string, error) {
	params := map[string]string{execField: execKey}
	url := coreHTTP.BuildURL(project, "tasks", "", params)
	b, status, err := coreHTTP.Do(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("lookup task failed (status %d): %w", status, err)
	}
	var tasks struct {
		Content []map[string]any `json:"content"`
	}
	if err := json.Unmarshal(b, &tasks); err != nil {
		return "", fmt.Errorf("invalid task lookup response: %w", err)
	}
	if tasks.Content == nil {
		return "", errors.New("invalid task lookup response: missing content array")
	}
	for _, task := range tasks.Content {
		if kind, _ := task["kind"].(string); kind == taskKind {
			if id, ok := task["id"]; ok && id != nil && fmt.Sprint(id) != "" {
				return fmt.Sprintf("%s://%s/%v", kind, project, id), nil
			}
			return "", errors.New("invalid task lookup response: matching task has no id")
		}
	}

	createURL := coreHTTP.BuildURL(project, "tasks", "", nil)
	body := map[string]any{
		"kind":    taskKind,
		"project": project,
		"spec": map[string]any{
			execField: execKey,
		},
	}
	data, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	respBytes, status, err := coreHTTP.Do(ctx, "POST", createURL, data)
	if err != nil {
		return "", fmt.Errorf("create task failed (status %d): %w", status, err)
	}

	var m map[string]any
	if err := json.Unmarshal(respBytes, &m); err != nil {
		return "", err
	}
	first, err := utils.GetFirstIfList(m)
	if err != nil {
		return "", err
	}

	k, okk := first["kind"].(string)
	idVal, oki := first["id"]
	if !okk || !oki {
		return "", errors.New("unable to obtain task key")
	}
	return fmt.Sprintf("%s://%s/%v", k, project, idVal), nil
}

// PrintMutationOutput prints the response from create, update, or run according to format ("short", "json", "yaml").
func PrintMutationOutput(respBody []byte, output string, successMessage string) error {
	format := utils.TranslateFormat(output)
	switch format {
	case "json":
		if len(respBody) == 0 {
			fmt.Println("{}")
			return nil
		}
		return printJson("id", respBody)
	case "yaml":
		if len(respBody) == 0 {
			return nil
		}
		return printYaml("id", respBody)
	default:
		if len(respBody) > 0 {
			_ = printShort(respBody)
		}
		return nil
	}
}
