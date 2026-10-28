// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package environment

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"dhcli/templates"
)

const agentDocsBase = "https://raw.githubusercontent.com/scc-digitalhub/docs"

var coreVersionPattern = regexp.MustCompile(`^v?([0-9]+)\.([0-9]+)(\.[0-9]+([A-Za-z0-9.+-]*))?$`)

type agentWorkspace struct {
	root    string
	version string
	branch  string
	docsURL string
	files   map[string]string
}

func docsBranch(version string) (string, error) {
	match := coreVersionPattern.FindStringSubmatch(strings.TrimSpace(version))
	if match == nil {
		return "", fmt.Errorf("invalid or missing Core version %q; cannot select a docs branch", version)
	}
	return match[1] + "." + match[2], nil
}

func fetchCoreConfiguration(client *http.Client, url string) ([]byte, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch %s: %s", url, resp.Status)
	}
	const maxSize = 8 << 20
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxSize+1))
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", url, err)
	}
	if len(body) == 0 || len(body) > maxSize {
		return nil, fmt.Errorf("fetch %s: empty response or exceeds %d bytes", url, maxSize)
	}
	return body, nil
}

func prepareAgentWorkspace(client *http.Client, endpoint, root string) (*agentWorkspace, error) {
	if endpoint == "" {
		return nil, fmt.Errorf("Core endpoint is missing; register and select an environment first")
	}
	body, err := fetchCoreConfiguration(client, strings.TrimRight(endpoint, "/")+"/.well-known/configuration")
	if err != nil {
		return nil, err
	}
	var config struct {
		Version string `json:"dhcore_version"`
	}
	if err := json.Unmarshal(body, &config); err != nil {
		return nil, fmt.Errorf("decode Core configuration: %w", err)
	}
	branch, err := docsBranch(config.Version)
	if err != nil {
		return nil, err
	}
	docsURL := agentDocsBase + "/" + branch
	files, err := templates.Render(docsURL)
	if err != nil {
		return nil, fmt.Errorf("render embedded agent templates: %w", err)
	}
	root, err = filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve agent workspace: %w", err)
	}
	workspace := &agentWorkspace{
		root:    root,
		version: config.Version,
		branch:  branch,
		docsURL: docsURL,
		files:   files,
	}
	if _, err := workspace.pendingFiles(); err != nil {
		return nil, err
	}
	return workspace, nil
}

func (w *agentWorkspace) paths() []string {
	paths := make([]string, 0, len(w.files))
	for path := range w.files {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths
}

func (w *agentWorkspace) pendingFiles() ([]string, error) {
	var pending []string
	for _, path := range w.paths() {
		target := filepath.Join(w.root, filepath.FromSlash(path))
		for parent := filepath.Dir(target); parent != w.root; parent = filepath.Dir(parent) {
			info, err := os.Lstat(parent)
			if os.IsNotExist(err) {
				continue
			}
			if err != nil {
				return nil, err
			}
			if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
				return nil, fmt.Errorf("agent destination is not a regular directory: %s", parent)
			}
		}
		info, err := os.Lstat(target)
		if os.IsNotExist(err) {
			pending = append(pending, path)
			continue
		}
		if err != nil {
			return nil, err
		}
		if info.Mode().IsRegular() {
			content, err := os.ReadFile(target)
			if err != nil {
				return nil, err
			}
			if string(content) == w.files[path] {
				continue
			}
		}
		return nil, fmt.Errorf("refusing to overwrite existing agent file %s; move it aside before initializing", target)
	}
	return pending, nil
}

func (w *agentWorkspace) install() (err error) {
	root, err := os.OpenRoot(w.root)
	if err != nil {
		return fmt.Errorf("open agent workspace: %w", err)
	}
	defer func() {
		if closeErr := root.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close agent workspace: %w", closeErr))
		}
	}()
	pending, err := w.pendingFiles()
	if err != nil {
		return err
	}
	var createdFiles, createdDirs []string
	defer func() {
		if err == nil {
			return
		}
		for i := len(createdFiles) - 1; i >= 0; i-- {
			if removeErr := root.Remove(createdFiles[i]); removeErr != nil {
				err = errors.Join(err, fmt.Errorf("rollback agent file: %w", removeErr))
			}
		}
		for i := len(createdDirs) - 1; i >= 0; i-- {
			if removeErr := root.Remove(createdDirs[i]); removeErr != nil {
				err = errors.Join(err, fmt.Errorf("rollback agent directory: %w", removeErr))
			}
		}
	}()
	for _, path := range pending {
		target := filepath.FromSlash(path)
		var missing []string
		for parent := filepath.Dir(target); parent != "."; parent = filepath.Dir(parent) {
			if _, statErr := root.Lstat(parent); os.IsNotExist(statErr) {
				missing = append(missing, parent)
			} else if statErr != nil {
				return fmt.Errorf("inspect agent directory: %w", statErr)
			} else {
				break
			}
		}
		for i := len(missing) - 1; i >= 0; i-- {
			if err := root.Mkdir(missing[i], 0755); err != nil {
				return fmt.Errorf("create agent directory: %w", err)
			}
			createdDirs = append(createdDirs, missing[i])
		}
		file, err := root.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
		if err != nil {
			return fmt.Errorf("create agent file %s: %w", target, err)
		}
		createdFiles = append(createdFiles, target)
		_, writeErr := file.WriteString(w.files[path])
		closeErr := file.Close()
		if writeErr != nil {
			return fmt.Errorf("write agent file %s: %w", target, writeErr)
		}
		if closeErr != nil {
			return fmt.Errorf("close agent file %s: %w", target, closeErr)
		}
	}
	return nil
}
