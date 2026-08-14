// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

	"dhcli/handlers/utils"

	"github.com/spf13/viper"
	"sigs.k8s.io/yaml"
)

const metricsFollowInterval = 15 * time.Second

// MetricsHandler fetches resource metrics from the API.
// scope must be one of: "instance", "project", "run".
// project is required for "project" and "run" scopes.
// id is required for the "run" scope.
func MetricsHandler(env string, output string, project string, scope string, id string, follow bool) error {
	utils.CheckUpdateEnvironment()
	utils.CheckApiLevel(utils.ApiLevelKey, utils.MetricsMin, utils.MetricsMax)
	if err := utils.CheckCredentials(); err != nil {
		return err
	}

	format := utils.TranslateFormat(output)

	baseURL := strings.TrimRight(viper.GetString(utils.DhCoreEndpoint), "/")
	apiVersion := viper.GetString(utils.DhCoreApiVersion)
	accessToken := viper.GetString(utils.DhCoreAccessToken)

	var metricsURL string
	switch scope {
	case "instance":
		metricsURL = fmt.Sprintf("%s/api/%s/resource_metrics", baseURL, apiVersion)
	case "project":
		if project == "" {
			return errors.New("project name is required")
		}
		metricsURL = fmt.Sprintf("%s/api/%s/projects/%s/resource_metrics", baseURL, apiVersion, project)
	case "run":
		if project == "" {
			return errors.New("project is required for run metrics (-p)")
		}
		if id == "" {
			return errors.New("run id is required")
		}
		metricsURL = fmt.Sprintf("%s/api/%s/-/%s/runs/%s/resource_metrics", baseURL, apiVersion, project, id)
	default:
		return fmt.Errorf("unknown scope: %s", scope)
	}

	httpClient := utils.GetDebugHTTPClient()
	if httpClient == nil {
		httpClient = &http.Client{}
	}

	first := true
	for {
		body, err := fetchMetrics(httpClient, metricsURL, accessToken)
		if err != nil {
			return err
		}

		switch format {
		case "json":
			fmt.Println(string(body))
		case "yaml":
			out, err := yaml.JSONToYAML(body)
			if err != nil {
				return fmt.Errorf("failed to convert to YAML: %w", err)
			}
			fmt.Print(string(out))
		default:
			if follow && !first {
				// Clear screen and move cursor to top for watch-like refresh
				fmt.Print("\033[2J\033[H")
			}
			first = false

			// Print entity header
			switch scope {
			case "instance":
				name := viper.GetString(utils.DhCoreName)
				if name == "" {
					name = baseURL
				}
				fmt.Printf("Instance: %s\n", name)
			case "project":
				fmt.Printf("Project: %s\n", project)
			case "run":
				fmt.Printf("Run: %s  Project: %s\n", id, project)
			}
			if follow {
				fmt.Printf("Updated: %s  (refresh every %s)\n", time.Now().Format("15:04:05"), metricsFollowInterval)
			}
			fmt.Println()

			if err := printMetricsShort(body); err != nil {
				return err
			}
		}

		if !follow {
			return nil
		}

		time.Sleep(metricsFollowInterval)
	}
}

func fetchMetrics(httpClient *http.Client, metricsURL string, accessToken string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, metricsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("server returned %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

type metricSummaryEntry struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

type metricDataPoint struct {
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
}

type metricEntry struct {
	Name    string               `json:"name"`
	Unit    string               `json:"unit"`
	Metrics []metricDataPoint    `json:"metrics"`
	Summary []metricSummaryEntry `json:"summary"`
}

type metricsResponse struct {
	Metrics []metricEntry `json:"metrics"`
}

func printMetricsShort(body []byte) error {
	var r metricsResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return fmt.Errorf("failed to parse metrics response: %w", err)
	}

	if len(r.Metrics) == 0 {
		fmt.Println("No metrics available.")
		return nil
	}

	// Find longest metric name for column alignment
	maxLen := 0
	for _, m := range r.Metrics {
		if len(m.Name) > maxLen {
			maxLen = len(m.Name)
		}
	}

	for _, m := range r.Metrics {
		var parts []string
		if len(m.Summary) > 0 {
			// avg first, then the rest in original order
			for _, s := range m.Summary {
				if s.Name == "avg" {
					parts = append(parts, fmt.Sprintf("avg=%s", formatMetricValue(m.Name, s.Value, m.Unit)))
					break
				}
			}
			for _, s := range m.Summary {
				if s.Name != "avg" {
					parts = append(parts, fmt.Sprintf("%s=%s", s.Name, formatMetricValue(m.Name, s.Value, m.Unit)))
				}
			}
		} else if len(m.Metrics) > 0 {
			v := m.Metrics[len(m.Metrics)-1].Value
			parts = append(parts, fmt.Sprintf("last=%s", formatMetricValue(m.Name, v, m.Unit)))
		}

		if len(parts) == 0 {
			parts = []string{"(no data)"}
		}

		fmt.Printf("%-*s : %s\n", maxLen, m.Name, strings.Join(parts, "  "))
	}

	return nil
}

// formatMetricValue formats a metric value according to its name and unit,
// mirroring the console's formatMetricsValue logic.
func formatMetricValue(name string, value float64, unit string) string {
	nameLower := strings.ToLower(name)

	// Apply default unit by name when unit is missing
	effectiveUnit := unit
	if effectiveUnit == "" {
		switch {
		case strings.Contains(nameLower, "cpu"):
			effectiveUnit = "n"
		case strings.Contains(nameLower, "memory"), strings.Contains(nameLower, "disk"):
			effectiveUnit = "bytes"
		}
	}

	switch effectiveUnit {
	case "n": // nanocores → millicores → cores
		millis := math.Floor(value * 1000)
		if millis >= 1000 {
			return trim2(millis / 1000)
		}
		return fmt.Sprintf("%sm", trim2(millis))

	case "m": // millicores → cores
		millis := math.Floor(value)
		if millis >= 1000 {
			return trim2(millis / 1000)
		}
		return fmt.Sprintf("%sm", trim2(millis))

	case "seconds":
		return trim2(value)

	case "bytes", "B":
		return prettyBytes(math.Floor(value))

	case "megabytes", "Mb", "MB":
		return prettyBytes(math.Floor(value * 1024 * 1024))

	case "percent":
		return fmt.Sprintf("%d%%", int(math.Floor(value)))

	default:
		return fmt.Sprintf("%d", int(math.Floor(value)))
	}
}

// trim2 formats a float with up to 2 decimal places, stripping trailing zeros.
func trim2(v float64) string {
	s := fmt.Sprintf("%.2f", v)
	if strings.Contains(s, ".") {
		s = strings.TrimRight(s, "0")
		s = strings.TrimRight(s, ".")
	}
	return s
}

// prettyBytes converts bytes to a human-readable binary string (KiB, MiB, GiB…).
func prettyBytes(b float64) string {
	units := []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB"}
	i := 0
	for b >= 1024 && i < len(units)-1 {
		b /= 1024
		i++
	}
	if i == 0 {
		return fmt.Sprintf("%dB", int(b))
	}
	return fmt.Sprintf("%s%s", trim2(b), units[i])
}
