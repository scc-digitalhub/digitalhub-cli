// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package templates

import (
	"embed"
	"fmt"
	"io/fs"
	"strings"
)

//go:embed AGENTS.md CLAUDE.md skills
var files embed.FS

// Render returns workspace-relative files with version-pinned knowledge references.
func Render(docsURL string) (map[string]string, error) {
	rendered := make(map[string]string)
	err := fs.WalkDir(files, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		content, err := files.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read embedded template %s: %w", path, err)
		}
		target := path
		if strings.HasPrefix(path, "skills/") {
			target = ".agents/" + path
		}
		rendered[target] = strings.ReplaceAll(string(content), "{{DOCS}}", docsURL)
		return nil
	})
	return rendered, err
}
