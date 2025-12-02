// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"bytes"
	"context"
	"dhcli/sdk/config"
	"dhcli/sdk/services/transfer"
	"dhcli/sdk/utils"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"sigs.k8s.io/yaml"
)

func DownloadHandler(env string, destination string, output string, project string, name string, resource string, id string, verbose bool) error {

	utils.CheckUpdateEnvironment()
	utils.CheckApiLevel(utils.ApiLevelKey, utils.LoginMin, utils.LoginMax)

	endpoint := utils.TranslateEndpoint(resource)
	if endpoint != "projects" && project == "" {
		return errors.New("project is mandatory for non-project resources")
	}

	// Traduce viper -> sdk.Config (INI/ENV/flags già caricati nel PersistentPreRunE)
	cfg := config.Config{
		Core: config.CoreConfig{
			BaseURL:     viper.GetString(utils.DhCoreEndpoint),
			APIVersion:  viper.GetString(utils.DhCoreApiVersion),
			AccessToken: viper.GetString(utils.DhCoreAccessToken),
		},
		S3: config.S3Config{
			AccessKey:   viper.GetString("aws_access_key_id"),
			SecretKey:   viper.GetString("aws_secret_access_key"),
			AccessToken: viper.GetString("aws_session_token"),
			Region:      viper.GetString("aws_region"),
			EndpointURL: viper.GetString("aws_endpoint_url"),
		},
	}

	svc, err := transfer.NewTransferService(context.Background(), cfg)
	if err != nil {
		return fmt.Errorf("sdk init failed: %w", err)
	}

	infos, err := svc.Download(context.Background(), endpoint, transfer.DownloadRequest{
		Project:     project,
		Resource:    resource,
		ID:          id,
		Name:        name,
		Destination: destination,
		Verbose:     verbose,
	})
	if err != nil {
		return err
	}

	switch utils.TranslateFormat(output) {
	case "json":
		return printDownloadJSON(infos)
	case "yaml":
		return printDownloadYAML(infos)
	default:
		printDownloadShort(infos)
		return nil
	}
}

// --- output identico a prima ---

func printDownloadShort(items []transfer.DownloadInfo) {
	for _, it := range items {
		fmt.Println(it.Path)
	}
}

func printDownloadJSON(items []transfer.DownloadInfo) error {
	data, err := json.Marshal(items)
	if err != nil {
		return err
	}
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, data, "", "    "); err != nil {
		return err
	}
	fmt.Println(pretty.String())
	return nil
}

func printDownloadYAML(items []transfer.DownloadInfo) error {
	out, err := yaml.Marshal(items)
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}

// (opzionale) utile nel caso serva in futuro per comporre path locali
func dirBaseForLocalTarget(localPath string) string {
	clean := filepath.Clean(localPath)
	parent := filepath.Dir(clean)
	if parent == "." || parent == string(os.PathSeparator) {
		return ""
	}
	return parent
}
