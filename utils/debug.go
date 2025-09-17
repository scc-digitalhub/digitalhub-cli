package utils

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

// DebugPrintRequest prints an HTTP request explicitly (no httputil.DumpRequestOut).
// It redacts Authorization and safely peeks at the body without consuming it.
func DebugPrintRequest(req *http.Request) {
	const maxBody = 4096 // bytes to print at most

	u := req.URL
	if u == nil {
		u = &url.URL{}
	}

	// Build request line and URL bits
	fmt.Println("\n--- HTTP REQUEST ---")
	fmt.Printf("Method:   %s\n", req.Method)
	fmt.Printf("URL:      %s\n", u.String())
	fmt.Printf("Scheme:   %s\n", u.Scheme)
	fmt.Printf("Host:     %s\n", hostOr(u.Host, req.Host))
	fmt.Printf("Path:     %s\n", emptyDash(u.Path))
	if q := u.Query().Encode(); q != "" {
		fmt.Printf("Query:    %s\n", q)
	} else {
		fmt.Printf("Query:    -\n")
	}

	// Headers (sorted, with Authorization redacted)
	fmt.Println("Headers:")
	printHeadersRedacted(req.Header)

	// Body (peek & restore)
	if req.Body == nil {
		fmt.Println("Body:     -")
		fmt.Println("--------------------")
		return
	}
	var buf bytes.Buffer
	tee := io.TeeReader(req.Body, &buf)
	preview := make([]byte, maxBody)
	n, _ := io.ReadFull(tee, preview)
	if n <= 0 && buf.Len() == 0 {
		fmt.Println("Body:     (empty)")
		_ = req.Body.Close()
		req.Body = io.NopCloser(bytes.NewReader(nil))
		fmt.Println("--------------------")
		return
	}

	// Rewind original body (remaining + what we already read)
	rest, _ := io.ReadAll(tee) // read the rest after preview
	all := append(buf.Bytes(), rest...)
	req.Body.Close()
	req.Body = io.NopCloser(bytes.NewReader(all))

	// Print preview (truncated)
	if len(all) > maxBody {
		fmt.Printf("Body:     (%d bytes, showing first %d)\n", len(all), maxBody)
		fmt.Println(string(all[:maxBody]))
		fmt.Println("... [truncated]")
	} else {
		fmt.Printf("Body:     (%d bytes)\n", len(all))
		fmt.Println(string(all))
	}
	fmt.Println("--------------------")
}

func printHeadersRedacted(h http.Header) {
	if len(h) == 0 {
		fmt.Println("  -")
		return
	}
	keys := make([]string, 0, len(h))
	for k := range h {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		vals := append([]string(nil), h.Values(k)...)
		if strings.EqualFold(k, "Authorization") {
			for i := range vals {
				vals[i] = redactAuth(vals[i])
			}
		}
		fmt.Printf("  %s: %s\n", k, strings.Join(vals, ", "))
	}
}

func redactAuth(v string) string {
	parts := strings.SplitN(v, " ", 2)
	if len(parts) == 2 {
		return parts[0] + " : " + parts[1]
	}
	return "***REDACTED***"
}

func hostOr(a, b string) string {
	if a != "" {
		return a
	}
	if b != "" {
		return b
	}
	return "-"
}

func emptyDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
