// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package s3client

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Client struct {
	s3 *s3.Client
}

type Config struct {
	AccessKey   string
	SecretKey   string
	AccessToken string
	Region      string
	EndpointURL string
}

func NewClient(ctx context.Context, cfgCreds Config) (*Client, error) {
	creds := aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(
		cfgCreds.AccessKey,
		cfgCreds.SecretKey,
		cfgCreds.AccessToken,
	))

	// Load AWS configuration with credentials and region
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(creds),
		config.WithRegion(cfgCreds.Region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Initialize S3 client options
	s3Options := func(o *s3.Options) {
		// If a custom endpoint is provided, set it as BaseEndpoint
		if cfgCreds.EndpointURL != "" {
			o.BaseEndpoint = aws.String(cfgCreds.EndpointURL)
			o.UsePathStyle = true // Necessary for some S3-compatible services
		}
	}

	// Create S3 client with the specified options
	return &Client{
		s3: s3.NewFromConfig(cfg, s3Options),
	}, nil
}

type S3File struct {
	Path         string
	Name         string
	Size         int64
	LastModified string
}

// ListFiles lists all objects under a given prefix (like a folder)
// TODO add Continuation Token
func (c *Client) ListFiles(ctx context.Context, bucket string, prefix string, maxKeys *int32) ([]S3File, error) {
	input := &s3.ListObjectsV2Input{
		Bucket:  aws.String(bucket),
		Prefix:  aws.String(prefix),
		MaxKeys: maxKeys,
	}

	resp, err := c.s3.ListObjectsV2(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to list objects in S3: %w", err)
	}

	files := make([]S3File, 0, len(resp.Contents))
	for _, obj := range resp.Contents {
		name := *obj.Key
		if prefix != "" && strings.HasPrefix(name, prefix) {
			name = strings.TrimPrefix(name, prefix)
		}
		files = append(files, S3File{
			Path:         *obj.Key,
			Name:         name,
			Size:         *obj.Size,
			LastModified: obj.LastModified.Format("2025-06-02T15:04:05Z07:00"),
		})
	}

	return files, nil
}

// DownloadFile downloads a file from S3 and saves it locally
func (c *Client) DownloadFile(ctx context.Context, bucket, key, localPath string) error {

	fmt.Printf("📥 Downloading from S3 path: s3://%s/%s\n", bucket, key)
	fmt.Printf("Saving to local path: %s\n", localPath)

	output, err := c.s3.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err != nil {
		return fmt.Errorf("failed to get object from S3: %w", err)
	}
	defer output.Body.Close()

	file, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, output.Body)
	if err != nil {
		return fmt.Errorf("failed to write to local file: %w", err)
	}

	return nil
}

// UploadFile choose between normal upload or multipart based on threshold
func (c *Client) UploadFile(ctx context.Context, bucket, key string, file *os.File) (interface{}, error) {
	const threshold = 100 * 1024 * 1024

	// Get file info
	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat error: %w", err)
	}
	size := info.Size()
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("seek error: %w", err)
	}

	// Detect MIME TYPE
	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	mime := http.DetectContentType(buf[:n])
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("rewind error: %w", err)
	}

	fmt.Printf("📤 Uploading to s3://%s/%s (%.2fMB, type: %s)\n", bucket, key, float64(size)/(1024*1024), mime)

	// Multipart upload with manager
	if size > threshold {
		return manager.NewUploader(c.s3).Upload(ctx, &s3.PutObjectInput{
			Bucket:      aws.String(bucket),
			Key:         aws.String(key),
			Body:        file,
			ContentType: aws.String(mime),
		})
	}

	return c.s3.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(key),
		Body:          file,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(mime),
	})
}
