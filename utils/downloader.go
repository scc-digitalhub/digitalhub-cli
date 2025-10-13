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
	"time"

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

/* ------------ S3: file o directory (with continuation token) ------------ */

func DownloadS3FileOrDir(
	s3Client *s3client.Client,
	ctx context.Context,
	parsedPath *ParsedPath,
	localPath string,
	verbose bool,
) error {
	bucket := parsedPath.Host
	// normalizza: rimuovi eventuale leading "/" (alcuni artifact salvano "/xxx/..")
	path := strings.TrimPrefix(parsedPath.Path, "/")

	// Directory?
	if strings.HasSuffix(path, "/") {
		localBase := cleanLocalPath(localPath)

		var totalFiles int
		var totalBytes int64

		all, err := s3Client.ListFilesAll(ctx, bucket, path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Listing failed, proceeding without totals: %v\n", err)
		} else {
			totalFiles = len(all)
			for _, f := range all {
				totalBytes += f.Size
			}
			fmt.Fprintf(os.Stderr, "Found %d files (%.2f MB) under s3://%s/%s\n",
				totalFiles, float64(totalBytes)/(1024*1024), bucket, path)
		}
		fmt.Fprintf(os.Stderr, "Your download will start in few seconds...\n")

		// Scarica via WalkPrefix (pagination)
		pageSize := int32(1000)
		var idx int

		return s3Client.WalkPrefix(ctx, bucket, path, pageSize, func(obj s3types.Object) error {
			idx++
			key := aws.ToString(obj.Key)
			relativePath := strings.TrimPrefix(key, path)
			targetPath := filepath.Join(localBase, relativePath)

			if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
				return fmt.Errorf("failed to create local directory: %w", err)
			}

			if verbose {
				if totalFiles > 0 {
					fmt.Fprintf(os.Stderr, "[%d/%d] %s\n", idx, totalFiles, relativePath)
				} else {
					fmt.Fprintf(os.Stderr, "[%d] %s\n", idx, relativePath)
				}
			}

			if verbose {
				// barra di avanzamento
				hook := &s3client.ProgressHook{
					OnStart: func(k string, total int64) {
						if total > 0 {
							fmt.Fprintf(os.Stderr, "   └─ size: %.2f MB\n", float64(total)/(1024*1024))
						}
					},
					OnProgress: func(k string, written, total int64) {
						if total <= 0 {
							return
						}
						pct := float64(written) / float64(total) * 100
						fmt.Fprintf(os.Stderr, "\r   └─ downloading: %6.2f%%", pct)
					},
					OnDone: func(k string, total int64, took time.Duration) {
						if total > 0 {
							fmt.Fprintf(os.Stderr, "\r   └─ done:        100.00%% in %s\n", took.Truncate(100*time.Millisecond))
						} else {
							fmt.Fprintf(os.Stderr, "   └─ done in %s\n", took.Truncate(100*time.Millisecond))
						}
					},
				}
				if err := s3Client.DownloadFileWithProgress(ctx, bucket, key, targetPath, hook); err != nil {
					return fmt.Errorf("failed to download file: %w", err)
				}
			} else {
				// silenzioso
				if err := s3Client.DownloadFile(ctx, bucket, key, targetPath); err != nil {
					return fmt.Errorf("failed to download file: %w", err)
				}
			}

			return nil
		})
	}

	// Singolo file
	key := path
	fmt.Fprintf(os.Stderr, "Your download will start in few seconds...\n")
	if verbose {
		fmt.Fprintf(os.Stderr, "Downloading s3://%s/%s → %s\n", bucket, key, localPath)
		hook := &s3client.ProgressHook{
			OnStart: func(k string, total int64) {
				if total > 0 {
					fmt.Fprintf(os.Stderr, "   size: %.2f MB\n", float64(total)/(1024*1024))
				}
			},
			OnProgress: func(k string, written, total int64) {
				if total <= 0 {
					return
				}
				pct := float64(written) / float64(total) * 100
				fmt.Fprintf(os.Stderr, "\r   downloading: %6.2f%%", pct)
			},
			OnDone: func(k string, total int64, took time.Duration) {
				if total > 0 {
					fmt.Fprintf(os.Stderr, "\r   done:        100.00%% in %s\n", took.Truncate(100*time.Millisecond))
				} else {
					fmt.Fprintf(os.Stderr, "   done in %s\n", took.Truncate(100*time.Millisecond))
				}
			},
		}
		if err := s3Client.DownloadFileWithProgress(ctx, bucket, key, localPath, hook); err != nil {
			return fmt.Errorf("S3 download failed: %w", err)
		}
		return nil
	}

	// silenzioso
	if err := s3Client.DownloadFile(ctx, bucket, key, localPath); err != nil {
		return fmt.Errorf("S3 download failed: %w", err)
	}
	return nil
}

/* ------------ helpers ------------ */
func cleanLocalPath(path string) string {
	clean := filepath.Clean(path)
	parts := strings.Split(clean, string(os.PathSeparator))
	if len(parts) == 1 {
		return ""
	}
	return filepath.Join(parts[:len(parts)-1]...)
}
