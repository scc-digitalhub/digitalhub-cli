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

/* ------------ logging helpers (stderr) ------------ */

func infof(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "[INFO] "+format+"\n", a...)
}
func warnf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "[WARN] "+format+"\n", a...)
}

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
		if verbose {
			// in verbose provo a dare numeri totali prima di iniziare
			all, err := s3Client.ListFilesAll(ctx, bucket, path)
			if err != nil {
				warnf("Listing failed, proceeding without totals: %v", err)
			} else {
				totalFiles = len(all)
				for _, f := range all {
					totalBytes += f.Size
				}
				infof("Preparing download s3://%s/%s → %s (%d files, %.2f MB)",
					bucket, path, displayPath(localBase), totalFiles, float64(totalBytes)/(1024*1024))
			}
		}
		if !verbose || totalFiles == 0 {
			// messaggio minimo anche senza verbose (o se count non disponibile)
			infof("Preparing download s3://%s/%s → %s", bucket, path, displayPath(localBase))
		}

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
					fmt.Fprintf(os.Stderr, "   [%d/%d] %s\n", idx, totalFiles, relativePath)
				} else {
					fmt.Fprintf(os.Stderr, "   [%d] %s\n", idx, relativePath)
				}

				// barra di avanzamento
				hook := &s3client.ProgressHook{
					OnStart: func(k string, total int64) {
						if total > 0 {
							fmt.Fprintf(os.Stderr, "      └─ size: %.2f MB\n", float64(total)/(1024*1024))
						}
					},
					OnProgress: func(k string, written, total int64) {
						if total <= 0 {
							return
						}
						pct := float64(written) / float64(total) * 100
						fmt.Fprintf(os.Stderr, "\r      └─ downloading: %6.2f%%", pct)
					},
					OnDone: func(k string, total int64, took time.Duration) {
						if total > 0 {
							fmt.Fprintf(os.Stderr, "\r      └─ done:        100.00%% in %s\n", took.Truncate(100*time.Millisecond))
						} else {
							fmt.Fprintf(os.Stderr, "      └─ done in %s\n", took.Truncate(100*time.Millisecond))
						}
					},
				}
				if err := s3Client.DownloadFileWithProgress(ctx, bucket, key, targetPath, hook); err != nil {
					return fmt.Errorf("failed to download file: %w", err)
				}
			} else {
				// silenzioso dopo il banner iniziale
				if err := s3Client.DownloadFile(ctx, bucket, key, targetPath); err != nil {
					return fmt.Errorf("failed to download file: %w", err)
				}
			}

			return nil
		})
	}

	// Singolo file
	key := path
	if verbose {
		infof("Preparing download s3://%s/%s → %s", bucket, key, displayPath(localPath))
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

	// non-verbose: banner minimo + download silenzioso
	infof("Preparing download s3://%s/%s → %s", bucket, key, displayPath(localPath))
	if err := s3Client.DownloadFile(ctx, bucket, key, localPath); err != nil {
		return fmt.Errorf("S3 download failed: %w", err)
	}
	return nil
}

/* ------------ helpers ------------ */

// Rimuove l’ultimo segmento dal path locale in modo che i file della “cartella” S3
// vengano salvati senza includere il prefisso root.
func cleanLocalPath(path string) string {
	clean := filepath.Clean(path)
	parts := strings.Split(clean, string(os.PathSeparator))
	if len(parts) == 1 {
		return ""
	}
	return filepath.Join(parts[:len(parts)-1]...)
}

// per stampare cartelle vuote come "." invece di stringa vuota
func displayPath(p string) string {
	if p == "" {
		return "."
	}
	return p
}
