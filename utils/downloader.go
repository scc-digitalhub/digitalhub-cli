// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"context"
	s3client "dhcli/configs"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

/* ------------ HTTP ------------ */

func DownloadHTTPFile(url string, destination string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) { _ = Body.Close() }(resp.Body)

	out, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

/* ------------ S3: file o directory (paginato, continuation token) ------------ */

func DownloadS3FileOrDir(
	s3Client *s3client.Client,
	ctx context.Context,
	parsedPath *ParsedPath,
	localPath string,
) error {
	bucket := parsedPath.Host
	// normalizza: rimuovi leading "/" da path S3 (alcuni artifact salvano "/xxx/..")
	path := strings.TrimPrefix(parsedPath.Path, "/")

	// È una "cartella"?
	if strings.HasSuffix(path, "/") {
		// Comportamento "vecchio": NON includere il prefisso root nel path locale.
		// Salva sotto il parent di localPath (equivalente al tuo cleanLocalPath).
		localBase := cleanLocalPath(localPath)

		pageSize := int32(1000) // valore alto; se l'endpoint limita, WalkPrefix continua a paginare
		return s3Client.WalkPrefix(ctx, bucket, path, pageSize, func(obj s3types.Object) error {
			key := aws.ToString(obj.Key)

			// path relativo togliendo il prefix "root"
			relativePath := strings.TrimPrefix(key, path)
			targetPath := filepath.Join(localBase, relativePath)

			// crea directory destinazione se serve
			if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
				return fmt.Errorf("failed to create local directory: %w", err)
			}

			if err := s3Client.DownloadFile(ctx, bucket, key, targetPath); err != nil {
				return fmt.Errorf("failed to download file: %w", err)
			}
			return nil
		})
	}

	// Singolo file
	key := path
	if err := s3Client.DownloadFile(ctx, bucket, key, localPath); err != nil {
		return fmt.Errorf("S3 download failed: %w", err)
	}
	return nil
}

/* ------------ helpers ------------ */

// come il tuo vecchio cleanLocalPath: torna il parent dir di localPath
// (così il prefisso root S3 non è ricreato in locale)
func cleanLocalPath(path string) string {
	clean := filepath.Clean(path)
	parts := strings.Split(clean, string(os.PathSeparator))
	if len(parts) == 1 {
		return ""
	}
	return filepath.Join(parts[:len(parts)-1]...)
}
