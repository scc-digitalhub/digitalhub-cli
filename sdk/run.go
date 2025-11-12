// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package sdk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type RunRequest struct {
	Project      string
	TaskKind     string
	FunctionID   string
	FunctionName string
	InputSpec    map[string]interface{}

	// Allows adapter to pass the translated endpoint ("runs")
	ResolvedRunsEndpoint string
}

type RunService struct {
	http CoreHTTP
}

func NewRunService(ctx context.Context, conf Config) (*RunService, error) {
	if conf.Core.BaseURL == "" || conf.Core.APIVersion == "" {
		return nil, errors.New("invalid core config")
	}
	return &RunService{
		http: newHTTPCore(nil, conf.Core),
	}, nil
}

// taskToRunKind converts a task kind (e.g. "python+job", "python+job:task", "python+serve")
// into its corresponding run kind (e.g. "python+job:run", "python+serve:run").
func taskToRunKind(task string) string {
	task = strings.TrimSpace(task)
	if task == "" {
		return task
	}
	if i := strings.IndexByte(task, ':'); i >= 0 {
		// replace suffix with :run
		return task[:i] + ":run"
	}
	return task + ":run"
}

func (s *RunService) Run(ctx context.Context, req RunRequest) error {
	if req.Project == "" {
		return errors.New("project not specified")
	}
	if req.TaskKind == "" {
		return errors.New("task kind not specified")
	}

	// IMPORTANT: keep task handling EXACTLY as original (no normalization here)
	origTaskKind := req.TaskKind
	runKind := taskToRunKind(origTaskKind)

	// Resolve function (returns kind and key). We only need the key for spec.
	_, fnKey, err := s.resolveFunction(ctx, req.Project, req.FunctionID, req.FunctionName)
	if err != nil {
		return err
	}

	// Get or create the TASK using the ORIGINAL task kind (exact match),
	// identical to old behavior.
	taskKey, err := s.getTaskKey(ctx, req.Project, fnKey, origTaskKind)
	if err != nil {
		taskKey, err = s.createTask(ctx, req.Project, fnKey, origTaskKind)
		if err != nil {
			return err
		}
	}

	// Build spec: merge input, then enforce required fields
	spec := map[string]interface{}{}
	for k, v := range req.InputSpec {
		spec[k] = v
	}
	spec["task"] = taskKey
	spec["function"] = fnKey
	spec["local_execution"] = false

	// RUN body: ONLY change here vs original: use runKind (e.g. "python+job:run")
	body := map[string]interface{}{
		"kind":    runKind,
		"project": req.Project,
		"spec":    spec,
	}
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	// Pretty print RUN body (shows kind: "python+job:run", plus spec.task/function/local_execution)
	// if pretty, err := json.MarshalIndent(body, "", "  "); err == nil {
	// 	fmt.Println("───────────────────────────────────────────────")
	// 	fmt.Println("Run Request Body")
	// 	fmt.Println("───────────────────────────────────────────────")
	// 	fmt.Println(string(pretty))
	// 	fmt.Println("───────────────────────────────────────────────")
	// }

	endpoint := req.ResolvedRunsEndpoint
	if endpoint == "" {
		endpoint = "runs"
	}
	url := s.http.BuildURL(req.Project, endpoint, "", nil)
	fmt.Printf("POST %s\n", url)

	_, status, err := s.http.Do(ctx, "POST", url, data)
	if err != nil {
		return fmt.Errorf("run creation failed (status %d): %w", status, err)
	}
	return nil
}

func (s *RunService) resolveFunction(ctx context.Context, project, id, name string) (string, string, error) {
	var fn map[string]interface{}

	if id != "" {
		url := s.http.BuildURL(project, "functions", id, nil)
		b, status, err := s.http.Do(ctx, "GET", url, nil)
		if err != nil {
			return "", "", fmt.Errorf("get function by id failed (status %d): %w", status, err)
		}
		if err := json.Unmarshal(b, &fn); err != nil {
			return "", "", err
		}
	} else if name != "" {
		url := s.http.BuildURL(project, "functions", "", map[string]string{"name": name})
		b, status, err := s.http.Do(ctx, "GET", url, nil)
		if err != nil {
			return "", "", fmt.Errorf("get function by name failed (status %d): %w", status, err)
		}
		var m map[string]interface{}
		if err := json.Unmarshal(b, &m); err != nil {
			return "", "", err
		}
		first, err := getFirstIfList(m)
		if err != nil {
			return "", "", err
		}
		fn = first
	} else {
		return "", "", errors.New("you must provide the name or ID of the function to run")
	}

	kind, ok1 := fn["kind"].(string)
	idVal, ok2 := fn["id"]
	nameVal, ok3 := fn["name"]
	if !ok1 || !ok2 || !ok3 {
		return "", "", errors.New("unable to obtain function key")
	}
	fnName := fmt.Sprint(nameVal)
	fnID := fmt.Sprint(idVal)
	return kind, fmt.Sprintf("%s://%s/%s:%s", kind, project, fnName, fnID), nil
}

func (s *RunService) getTaskKey(ctx context.Context, project, functionKey, taskKind string) (string, error) {
	params := map[string]string{"function": functionKey}
	url := s.http.BuildURL(project, "tasks", "", params)
	b, _, err := s.http.Do(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		return "", err
	}

	if c, ok := m["content"].([]interface{}); ok {
		for _, it := range c {
			if tm, ok := it.(map[string]interface{}); ok {
				if k, ok := tm["kind"].(string); ok && k == taskKind {
					if idVal, ok := tm["id"]; ok {
						return fmt.Sprintf("%s://%s/%v", k, project, idVal), nil
					}
				}
			}
		}
	}
	return "", errors.New("unable to obtain task key")
}

func (s *RunService) createTask(ctx context.Context, project, functionKey, taskKind string) (string, error) {
	url := s.http.BuildURL(project, "tasks", "", nil)

	// EXACTLY like the old code: use the task kind AS-IS
	body := map[string]interface{}{
		"kind":    taskKind,
		"project": project,
		"spec": map[string]interface{}{
			"function": functionKey,
		},
	}
	data, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	// Debug print of TASK creation body (shows kind as passed, no normalization)
	// if pretty, err := json.MarshalIndent(body, "", "  "); err == nil {
	// 	fmt.Println("───────────────────────────────────────────────")
	// 	fmt.Println("Creating Task Body")
	// 	fmt.Println("───────────────────────────────────────────────")
	// 	fmt.Println(string(pretty))
	// 	fmt.Println("───────────────────────────────────────────────")
	// }

	b, status, err := s.http.Do(ctx, "POST", url, data)
	if err != nil {
		return "", fmt.Errorf("create task failed (status %d): %w", status, err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		return "", err
	}
	first, err := getFirstIfList(m)
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

// Minimal copy of utils.GetFirstIfList semantics for SDK isolation.
func getFirstIfList(m map[string]interface{}) (map[string]interface{}, error) {
	if c, ok := m["content"].([]interface{}); ok && len(c) > 0 {
		if mm, ok := c[0].(map[string]interface{}); ok {
			return mm, nil
		}
		return nil, errors.New("invalid content element")
	}
	return m, nil
}
