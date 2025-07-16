package utils

import (
	"context"
	s3client "dhcli/configs"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// UploadS3File uploads a single file and returns S3 response and metadata info
func UploadS3File(client *s3client.Client, ctx context.Context, bucket, key, localPath string) (*s3.PutObjectOutput, []map[string]interface{}, error) {
	file, err := os.Open(localPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open local file: %w", err)
	}
	defer file.Close()

	// Read for content-type detection
	header := make([]byte, 512)
	n, _ := file.Read(header)
	contentType := http.DetectContentType(header[:n])

	// Stat file for size and mod time
	info, err := file.Stat()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// Reset file reader before upload
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, nil, fmt.Errorf("failed to reset file reader: %w", err)
	}

	// Upload the file
	output, err := client.UploadFile(ctx, bucket, key, file)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to upload file to S3: %w", err)
	}

	// Prepare file info
	files := []map[string]interface{}{
		{
			"path":          "",
			"name":          info.Name(),
			"content_type":  contentType,
			"last_modified": info.ModTime().UTC().Format(time.RFC1123),
			"size":          info.Size(),
		},
	}

	return output, files, nil
}

// UploadS3Dir uploads all files inside a directory
func UploadS3Dir(client *s3client.Client, ctx context.Context, parsedPath *ParsedPath, localPath string) ([]*s3.PutObjectOutput, []map[string]interface{}, error) {
	bucket := parsedPath.Host
	prefix := parsedPath.Path

	var results []*s3.PutObjectOutput

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
			return fmt.Errorf("failed to compute relative path: %w", err)
		}

		s3Key := filepath.ToSlash(filepath.Join(prefix, relPath))

		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open file %s: %w", path, err)
		}
		defer file.Close()

		output, err := client.UploadFile(ctx, bucket, s3Key, file)
		if err != nil {
			return fmt.Errorf("failed to upload %s: %w", path, err)
		}

		results = append(results, output)

		dirPath := filepath.Dir(relPath)
		normalizedPath := ""
		if dirPath != "." {
			normalizedPath = filepath.ToSlash(dirPath)
		}

		// Re-open file to detect MIME type
		file.Seek(0, 0)
		buffer := make([]byte, 512)
		n, _ := file.Read(buffer)
		contentType := http.DetectContentType(buffer[:n])

		// Format last modified time
		lastModified := info.ModTime().UTC().Format(http.TimeFormat)

		// Collect file metadata
		fileInfos = append(fileInfos, map[string]interface{}{
			"path":          normalizedPath,
			"name":          info.Name(),
			"content_type":  contentType,
			"last_modified": lastModified,
			"size":          info.Size(),
		})

		return nil
	})

	if err != nil {
		return nil, nil, fmt.Errorf("failed to upload directory: %w", err)
	}

	return results, fileInfos, nil
}
