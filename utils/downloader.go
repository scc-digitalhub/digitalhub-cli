package utils

import (
	"context"
	s3client "dhcli/configs"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// DownloadHTTPFile function for get a file from http or https
func DownloadHTTPFile(url string, destination string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	out, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// DownloadS3FileOrDir Function for download file or directory form S3
func DownloadS3FileOrDir(s3Client *s3client.Client, ctx context.Context,
	parsedPath *ParsedPath, localPath string,
) error {
	bucket := parsedPath.Host
	path := parsedPath.Path

	//TODO for pagination use ContinuationToken (check how to do it)

	// If folder
	if strings.HasSuffix(path, "/") {

		localPath = cleanLocalPath(localPath)

		files, err := s3Client.ListFiles(ctx, bucket, path, aws.Int32(200)) //TODO remove maxKeys???
		if err != nil {
			return fmt.Errorf("failed to list S3 folder: %w", err)
		}

		for _, file := range files {
			// Build path
			// TODO strip prefix and keep all the folder structure
			relativePath := strings.TrimPrefix(file.Path, path)

			fmt.Printf("ParsedPath: Host=%s, Path=%s\n, LocalPath=%s\n", parsedPath.Host, parsedPath.Path, localPath)
			targetPath := filepath.Join(localPath, relativePath)

			// Create a directory if necessary
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return fmt.Errorf("failed to create local directory: %w", err)
			}

			if err := s3Client.DownloadFile(ctx, bucket, file.Path, targetPath); err != nil {
				return fmt.Errorf("failed to download file: %w", err)
			}
		}
	} else {
		// Single file
		if err := s3Client.DownloadFile(ctx, bucket, path, localPath); err != nil {
			return fmt.Errorf("S3 download failed: %w", err)
		}
	}

	return nil
}

func cleanLocalPath(path string) string {
	clean := filepath.Clean(path)
	parts := strings.Split(clean, string(os.PathSeparator))

	if len(parts) == 1 {
		return ""
	}
	return filepath.Join(parts[:len(parts)-1]...)
}
