package utils

import (
	"context"
	s3client "dhcli/configs"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"os"
	"path/filepath"
)

// UploadS3File uploads a single file to a specified S3 bucket and key.
// It returns the S3 PutObjectOutput containing metadata like ETag.
func UploadS3File(client *s3client.Client, ctx context.Context, bucket, key, localPath string) (*s3.PutObjectOutput, error) {
	file, err := os.Open(localPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open local file: %w", err)
	}
	defer file.Close()

	output, err := client.UploadFile(ctx, bucket, key, file)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file to S3: %w", err)
	}

	return output, nil
}

// UploadS3Dir uploads all files inside a local directory to an S3 path.
// It preserves the folder structure relative to the root of localPath.
func UploadS3Dir(client *s3client.Client, ctx context.Context, parsedPath *ParsedPath, localPath string) ([]*s3.PutObjectOutput, error) {
	bucket := parsedPath.Host
	prefix := parsedPath.Path

	var results []*s3.PutObjectOutput

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

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to upload directory: %w", err)
	}

	return results, nil
}
