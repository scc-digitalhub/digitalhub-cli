package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type CoreHTTP interface {
	BuildURL(project, resource, id string, params map[string]string) string
	Do(ctx context.Context, method, url string, data []byte) ([]byte, int, error)
}

type httpCore struct {
	h   *http.Client
	cfg CoreConfig
}

func newHTTPCore(h *http.Client, cfg CoreConfig) CoreHTTP {
	if h == nil {
		h = http.DefaultClient
	}
	return &httpCore{h: h, cfg: cfg}
}

func (a *httpCore) BuildURL(project, resource, id string, params map[string]string) string {
	base := fmt.Sprintf("%s/api/%s", a.cfg.BaseURL, a.cfg.APIVersion)
	if resource != "projects" && project != "" {
		base += "/-/" + project
	}
	base += "/" + resource
	if id != "" {
		base += "/" + id
	}
	first := true
	for k, v := range params {
		if v == "" {
			continue
		}
		if first {
			base += "?"
			first = false
		} else {
			base += "&"
		}
		base += fmt.Sprintf("%s=%s", k, v)
	}
	return base
}

func (a *httpCore) Do(ctx context.Context, method, url string, data []byte) ([]byte, int, error) {
	var body io.Reader
	if data != nil {
		body = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, 0, err
	}
	if data != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if tok := a.cfg.AccessToken; tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}

	resp, err := a.h.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	b, rerr := io.ReadAll(resp.Body)
	// Replica del comportamento originale: su non-200 includi "message"
	if resp.StatusCode != 200 {
		var m map[string]any
		if json.Unmarshal(b, &m) == nil {
			if msg, ok := m["message"].(string); ok && msg != "" {
				return b, resp.StatusCode, fmt.Errorf("core responded with: %s - %s", resp.Status, msg)
			}
		}
		return b, resp.StatusCode, fmt.Errorf("core responded with: %s", resp.Status)
	}
	return b, resp.StatusCode, rerr
}
