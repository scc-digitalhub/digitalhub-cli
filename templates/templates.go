// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package templates

import (
	"archive/tar"
	"compress/gzip"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"path"
	"strings"

	"github.com/cbroglie/mustache"
)

const (
	RemoteArchiveURL = "https://codeload.github.com/scc-digitalhub/digitalhub-cli/tar.gz/refs/heads/main"
	RemoteSourceRef  = "scc-digitalhub/digitalhub-cli@main"
)

//go:embed AGENTS.md CLAUDE.md skills
var files embed.FS

// SourceInfo describes where the agent templates were loaded from and their version/revision.
type SourceInfo struct {
	Kind        string
	Ref         string
	Revision    string
	FallbackErr error
}

// String returns a user-facing description of the template source and version.
func (s SourceInfo) String() string {
	if s.Kind == "remote" {
		if s.Revision != "" {
			return fmt.Sprintf("remote (%s, revision %s)", s.Ref, s.Revision)
		}
		return fmt.Sprintf("remote (%s)", s.Ref)
	}
	if s.FallbackErr != nil {
		return fmt.Sprintf("embedded (%s; remote fetch unavailable: %v)", s.Ref, s.FallbackErr)
	}
	return fmt.Sprintf("embedded (%s)", s.Ref)
}

// Render returns workspace-relative files from the embedded templates with version-pinned knowledge references.
func Render(docsURL string) (map[string]string, error) {
	raw, err := loadEmbeddedRaw()
	if err != nil {
		return nil, err
	}
	return renderRawMap(raw, docsURL)
}

// RenderLatest fetches the latest templates from the remote GitHub repository tarball and falls back
// to the embedded CLI templates if the remote archive cannot be fetched, validated, or rendered.
func RenderLatest(client *http.Client, docsURL string) (map[string]string, SourceInfo, error) {
	return RenderFromArchiveOrEmbedded(client, RemoteArchiveURL, docsURL)
}

// RenderFromArchiveOrEmbedded attempts to load and render templates from archiveURL,
// falling back to embedded templates if fetching, validating, or rendering the remote archive fails.
func RenderFromArchiveOrEmbedded(client *http.Client, archiveURL, docsURL string) (map[string]string, SourceInfo, error) {
	raw, rev, fetchErr := fetchRemoteRaw(client, archiveURL)
	if fetchErr == nil {
		rendered, renderErr := renderRawMap(raw, docsURL)
		if renderErr == nil {
			return rendered, SourceInfo{
				Kind:     "remote",
				Ref:      RemoteSourceRef,
				Revision: rev,
			}, nil
		}
		fetchErr = renderErr
	}

	embeddedRaw, err := loadEmbeddedRaw()
	if err != nil {
		return nil, SourceInfo{}, err
	}
	rendered, err := renderRawMap(embeddedRaw, docsURL)
	if err != nil {
		return nil, SourceInfo{}, err
	}
	return rendered, SourceInfo{
		Kind:        "embedded",
		Ref:         "built-in CLI templates",
		FallbackErr: fetchErr,
	}, nil
}

func fetchRemoteRaw(client *http.Client, archiveURL string) (map[string][]byte, string, error) {
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Get(archiveURL)
	if err != nil {
		return nil, "", fmt.Errorf("fetch %s: %w", archiveURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("fetch %s: %s", archiveURL, resp.Status)
	}

	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("decompress archive: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	raw := make(map[string][]byte)
	hasSkill := false
	var revision string

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, "", fmt.Errorf("read tar archive: %w", err)
		}

		if hdr.Typeflag == tar.TypeXGlobalHeader {
			if revision == "" {
				revision = shortRevision(hdr.PAXRecords["comment"])
			}
			continue
		}

		if hdr.Typeflag != tar.TypeReg && hdr.Typeflag != tar.TypeRegA {
			continue
		}

		relPath, ok := extractTemplateRelPath(hdr.Name)
		if !ok {
			continue
		}

		content, err := io.ReadAll(tr)
		if err != nil {
			return nil, "", fmt.Errorf("read remote template %s: %w", relPath, err)
		}
		raw[relPath] = content
		if strings.HasPrefix(relPath, "skills/") && strings.Count(relPath, "/") == 2 && path.Base(relPath) == "SKILL.md" {
			hasSkill = true
		}
	}

	for _, required := range []string{"AGENTS.md", "CLAUDE.md"} {
		if _, ok := raw[required]; !ok {
			return nil, "", fmt.Errorf("remote archive missing required %s", required)
		}
	}
	if !hasSkill {
		return nil, "", fmt.Errorf("remote archive missing required skills/<name>/SKILL.md template")
	}

	return raw, revision, nil
}

func extractTemplateRelPath(entryName string) (string, bool) {
	if strings.Contains(entryName, `\`) {
		return "", false
	}
	trimmed := strings.TrimLeft(entryName, "/")
	if trimmed == "" {
		return "", false
	}
	clean := path.Clean(trimmed)
	if clean == "." || strings.HasPrefix(clean, "../") || strings.Contains(clean, "/../") {
		return "", false
	}
	// Strip the top-level GitHub archive directory (e.g. "digitalhub-cli-main/").
	_, afterRoot, ok := strings.Cut(clean, "/")
	if !ok {
		return "", false
	}
	relPath, ok := strings.CutPrefix(afterRoot, "templates/")
	if !ok || relPath == "" {
		return "", false
	}
	if !isAllowedTemplatePath(relPath) {
		return "", false
	}
	return relPath, true
}

func isAllowedTemplatePath(relPath string) bool {
	if relPath == "AGENTS.md" || relPath == "CLAUDE.md" {
		return true
	}
	return strings.HasPrefix(relPath, "skills/") && !strings.HasSuffix(relPath, ".go")
}

func shortRevision(rev string) string {
	rev = strings.TrimSpace(rev)
	if len(rev) >= 7 && isHex(rev) {
		if len(rev) > 7 {
			return rev[:7]
		}
		return rev
	}
	return ""
}

func isHex(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return len(s) > 0
}

func loadEmbeddedRaw() (map[string][]byte, error) {
	raw := make(map[string][]byte)
	err := fs.WalkDir(files, ".", func(filePath string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		content, err := files.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("read embedded template %s: %w", filePath, err)
		}
		raw[filePath] = content
		return nil
	})
	return raw, err
}

func renderRawMap(raw map[string][]byte, docsURL string) (map[string]string, error) {
	rendered := make(map[string]string, len(raw))
	for relPath, content := range raw {
		target := relPath
		if strings.HasPrefix(relPath, "skills/") {
			target = ".agents/" + relPath
		}
		result, err := mustache.RenderRaw("{{=<% %>=}}\n"+string(content), true, map[string]string{"DOCS": docsURL})
		if err != nil {
			return nil, fmt.Errorf("render template %s: %w", relPath, err)
		}
		result = strings.ReplaceAll(result, "{{DOCS}}", docsURL)
		rendered[target] = result
	}
	return rendered, nil
}
