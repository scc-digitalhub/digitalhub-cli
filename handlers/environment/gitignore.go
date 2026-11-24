// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package environment

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func UpdateInitGitignore(paths []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("resolve current directory: %w", err)
	}
	escape := strings.NewReplacer(`\`, `\\`, "*", `\*`, "?", `\?`, "[", `\[`, "]", `\]`, " ", `\ `)
	var patterns []string
	for _, path := range paths {
		absolute, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("resolve path %q: %w", path, err)
		}
		relative, err := filepath.Rel(cwd, absolute)
		if err != nil || !filepath.IsLocal(relative) || relative == "." {
			log.Printf("Cannot ignore %q in the current directory's .gitignore; add an appropriate rule manually.", path)
			continue
		}
		relative = filepath.ToSlash(relative)
		if strings.ContainsAny(relative, "\r\n") {
			return fmt.Errorf("cannot add path containing a newline to .gitignore: %q", path)
		}
		info, err := os.Lstat(absolute)
		if err != nil {
			return fmt.Errorf("inspect initialized path %q: %w", path, err)
		}
		pattern := escape.Replace(relative)
		if strings.HasPrefix(pattern, "#") || strings.HasPrefix(pattern, "!") {
			pattern = `\` + pattern
		}
		if info.IsDir() {
			pattern += "/"
		}
		patterns = append(patterns, pattern)
	}
	const path = ".gitignore"
	info, err := os.Lstat(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("inspect .gitignore: %w", err)
	}
	if err == nil && !info.Mode().IsRegular() {
		return fmt.Errorf("refusing to update non-regular .gitignore")
	}
	content := []byte(nil)
	if err == nil {
		content, err = os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read .gitignore: %w", err)
		}
	}

	lines := strings.Split(string(content), "\n")
	var missing []string
	seen := make(map[string]bool)
	for _, pattern := range patterns {
		if seen[pattern] {
			continue
		}
		seen[pattern] = true
		found := false
		for _, line := range lines {
			line = strings.TrimSuffix(line, "\r")
			if line == pattern {
				found = true
			} else if strings.HasPrefix(line, "!") {
				found = false
			}
		}
		if !found {
			missing = append(missing, pattern)
		}
	}
	if len(missing) == 0 && info != nil {
		return nil
	}
	var addition string
	if len(missing) > 0 {
		if len(content) > 0 && content[len(content)-1] != '\n' {
			addition = "\n"
		}
		addition += strings.Join(missing, "\n") + "\n"
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("open .gitignore: %w", err)
	}
	if _, err := file.WriteString(addition); err != nil {
		file.Close()
		return fmt.Errorf("update .gitignore: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close .gitignore: %w", err)
	}
	return nil
}
