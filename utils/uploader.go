package utils

import (
	"context"
	s3client "dhcli/configs"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func UploadS3File(client *s3client.Client, ctx context.Context, bucket, key, localPath string) (map[string]interface{}, []map[string]interface{}, error) {
	file, err := os.Open(localPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open local file: %w", err)
	}
	defer file.Close()

	// Detect content-type
	header := make([]byte, 512)
	n, _ := file.Read(header)
	contentType := http.DetectContentType(header[:n])

	// File info
	info, err := file.Stat()
	if err != nil {
		return nil, nil, fmt.Errorf("stat error: %w", err)
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, nil, fmt.Errorf("seek error: %w", err)
	}

	// Upload
	output, err := client.UploadFile(ctx, bucket, key, file)
	if err != nil {
		return nil, nil, fmt.Errorf("upload error: %w", err)
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

func UploadS3Dir(client *s3client.Client, ctx context.Context, parsedPath *ParsedPath, localPath string) ([]map[string]interface{}, []map[string]interface{}, error) {
	bucket := parsedPath.Host
	prefix := parsedPath.Path

	var results []map[string]interface{}
	var fileInfos []map[string]interface{}

	err := filepath.Walk(localPath, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("walk error: %w", walkErr)
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(localPath, path)
		if err != nil {
			return fmt.Errorf("relative path error: %w", err)
		}
		s3Key := filepath.ToSlash(filepath.Join(prefix, relPath))

		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("open file error: %w", err)
		}
		defer file.Close()

		output, err := client.UploadFile(ctx, bucket, s3Key, file)
		if err != nil {
			return fmt.Errorf("upload error (%s): %w", path, err)
		}

		// Normalize result
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
		results = append(results, result)

		// MIME
		file.Seek(0, 0)
		buf := make([]byte, 512)
		n, _ := file.Read(buf)
		contentType := http.DetectContentType(buf[:n])

		dirPath := filepath.Dir(relPath)
		normalizedPath := ""
		if dirPath != "." {
			normalizedPath = filepath.ToSlash(dirPath)
		}

		fileInfos = append(fileInfos, map[string]interface{}{
			"path":          normalizedPath,
			"name":          info.Name(),
			"content_type":  contentType,
			"last_modified": info.ModTime().UTC().Format(http.TimeFormat),
			"size":          info.Size(),
		})

		return nil
	})

	if err != nil {
		return nil, nil, fmt.Errorf("failed to upload directory: %w", err)
	}

	return results, fileInfos, nil
}
