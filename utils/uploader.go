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
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

/* ------------ logging helpers (stderr) ------------ */

func upInfof(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "[INFO] "+format+"\n", a...)
}
func upWarnf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "[WARN] "+format+"\n", a...)
}

/* ------------ FILE SINGOLO ------------ */

func UploadS3File(client *s3client.Client, ctx context.Context, bucket, key, localPath string, verbose bool) (map[string]interface{}, []map[string]interface{}, error) {
	file, err := os.Open(localPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open local file: %w", err)
	}
	defer file.Close()

	// Detect content-type
	header := make([]byte, 512)
	n, _ := file.Read(header)
	contentType := http.DetectContentType(header[:n])
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, nil, fmt.Errorf("seek error: %w", err)
	}

	// Banner
	if verbose {
		upInfof("Preparing upload %s → s3://%s/%s", displayPathUpload(localPath), bucket, key)
	} else {
		upInfof("Preparing upload %s → s3://%s/%s", displayPathUpload(localPath), bucket, key)
	}

	// Upload
	var output interface{}
	if verbose {
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
				fmt.Fprintf(os.Stderr, "\r   uploading: %6.2f%%", pct)
			},
			OnDone: func(k string, total int64, took time.Duration) {
				if total > 0 {
					fmt.Fprintf(os.Stderr, "\r   done:      100.00%% in %s\n", took.Truncate(100*time.Millisecond))
				} else {
					fmt.Fprintf(os.Stderr, "   done in %s\n", took.Truncate(100*time.Millisecond))
				}
			},
		}
		// rewind per sicurezza
		if _, err := file.Seek(0, io.SeekStart); err != nil {
			return nil, nil, fmt.Errorf("seek error: %w", err)
		}
		output, err = client.UploadFileWithProgress(ctx, bucket, key, file, hook)
		if err != nil {
			return nil, nil, fmt.Errorf("upload error: %w", err)
		}
	} else {
		// rewind per sicurezza
		if _, err := file.Seek(0, io.SeekStart); err != nil {
			return nil, nil, fmt.Errorf("seek error: %w", err)
		}
		output, err = client.UploadFile(ctx, bucket, key, file)
		if err != nil {
			return nil, nil, fmt.Errorf("upload error: %w", err)
		}
	}

	// Normalize upload response to map
	result := map[string]interface{}{}
	switch v := output.(type) {
	case *s3.PutObjectOutput:
		if v.ETag != nil {
			result["etag"] = *v.ETag
		}
		if v.VersionId != nil {
			result["version_id"] = *v.VersionId
		}
	case *manager.UploadOutput:
		result["location"] = v.Location
		result["upload_id"] = v.UploadID
	}

	// Describe the uploaded file
	info, err := os.Stat(localPath)
	if err != nil {
		return result, nil, nil // fallback: response ok anche senza file info
	}

	files := []map[string]interface{}{
		{
			"path":          "",
			"name":          info.Name(),
			"content_type":  contentType,
			"last_modified": info.ModTime().UTC().Format(time.RFC1123),
			"size":          info.Size(),
		},
	}

	return result, files, nil
}

/* ------------ DIRECTORY ------------ */

func UploadS3Dir(client *s3client.Client, ctx context.Context, parsedPath *ParsedPath, localPath string, verbose bool) ([]map[string]interface{}, []map[string]interface{}, error) {
	bucket := parsedPath.Host
	prefix := parsedPath.Path

	// Enumerazione file locali (per poter stampare [i/N] e totals)
	var localFiles []string
	err := filepath.Walk(localPath, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("walk error: %w", walkErr)
		}
		if info.IsDir() {
			return nil
		}
		localFiles = append(localFiles, path)
		return nil
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to enumerate local directory: %w", err)
	}

	total := len(localFiles)
	if verbose {
		upInfof("Preparing upload directory %s → s3://%s/%s (%d files)", displayPathUpload(localPath), bucket, prefix, total)
	} else {
		upInfof("Preparing upload directory %s → s3://%s/%s", displayPathUpload(localPath), bucket, prefix)
	}

	var results []map[string]interface{}
	var fileInfos []map[string]interface{}

	for i, path := range localFiles {
		info, err := os.Stat(path)
		if err != nil {
			return nil, nil, fmt.Errorf("stat error on %s: %w", path, err)
		}
		relPath, err := filepath.Rel(localPath, path)
		if err != nil {
			return nil, nil, fmt.Errorf("relative path error: %w", err)
		}
		s3Key := filepath.ToSlash(filepath.Join(prefix, relPath))

		file, err := os.Open(path)
		if err != nil {
			return nil, nil, fmt.Errorf("open file error: %w", err)
		}

		// MIME
		header := make([]byte, 512)
		n, _ := file.Read(header)
		contentType := http.DetectContentType(header[:n])
		if _, err := file.Seek(0, io.SeekStart); err != nil {
			_ = file.Close()
			return nil, nil, fmt.Errorf("seek error: %w", err)
		}

		if verbose {
			fmt.Fprintf(os.Stderr, "   [%d/%d] %s → s3://%s/%s\n", i+1, total, relPath, bucket, s3Key)
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
					fmt.Fprintf(os.Stderr, "\r      └─ uploading: %6.2f%%", pct)
				},
				OnDone: func(k string, total int64, took time.Duration) {
					if total > 0 {
						fmt.Fprintf(os.Stderr, "\r      └─ done:      100.00%% in %s\n", took.Truncate(100*time.Millisecond))
					} else {
						fmt.Fprintf(os.Stderr, "      └─ done in %s\n", took.Truncate(100*time.Millisecond))
					}
				},
			}
			if _, err := file.Seek(0, io.SeekStart); err != nil {
				_ = file.Close()
				return nil, nil, fmt.Errorf("seek error: %w", err)
			}
			out, upErr := client.UploadFileWithProgress(ctx, bucket, s3Key, file, hook)
			_ = file.Close()
			if upErr != nil {
				return nil, nil, fmt.Errorf("upload error (%s): %w", path, upErr)
			}

			results = append(results, normalizeUploadResult(out))
		} else {
			if _, err := file.Seek(0, io.SeekStart); err != nil {
				_ = file.Close()
				return nil, nil, fmt.Errorf("seek error: %w", err)
			}
			out, upErr := client.UploadFile(ctx, bucket, s3Key, file)
			_ = file.Close()
			if upErr != nil {
				return nil, nil, fmt.Errorf("upload error (%s): %w", path, upErr)
			}
			results = append(results, normalizeUploadResult(out))
		}

		// Accumula info file per status
		dirPath := filepath.Dir(relPath)
		normalizedPath := info.Name()
		if dirPath != "." {
			normalizedPath = filepath.ToSlash(dirPath + "/" + info.Name())
		}
		fileInfos = append(fileInfos, map[string]interface{}{
			"path":          normalizedPath,
			"name":          info.Name(),
			"content_type":  contentType,
			"last_modified": info.ModTime().UTC().Format(http.TimeFormat),
			"size":          info.Size(),
		})
	}

	return results, fileInfos, nil
}

/* ------------ helpers ------------ */

func normalizeUploadResult(output interface{}) map[string]interface{} {
	result := map[string]interface{}{}
	switch v := output.(type) {
	case *s3.PutObjectOutput:
		if v.ETag != nil {
			result["etag"] = *v.ETag
		}
		if v.VersionId != nil {
			result["version_id"] = *v.VersionId
		}
	case *manager.UploadOutput:
		result["location"] = v.Location
		result["upload_id"] = v.UploadID
	}
	return result
}

func displayPathUpload(p string) string {
	if p == "" {
		return "."
	}
	return p
}
