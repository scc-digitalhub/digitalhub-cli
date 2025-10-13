// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package s3client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
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

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(creds),
		config.WithRegion(cfgCreds.Region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	s3Options := func(o *s3.Options) {
		if cfgCreds.EndpointURL != "" {
			o.BaseEndpoint = aws.String(cfgCreds.EndpointURL)
			o.UsePathStyle = true // necessario per molti S3-compat
		}
	}

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

/* -------------------- LIST (paginata) -------------------- */

func (c *Client) ListFilesPaged(
	ctx context.Context,
	bucket string,
	prefix string,
	maxKeys *int32,
	continuationToken *string,
) ([]S3File, *string, error) {
	input := &s3.ListObjectsV2Input{
		Bucket:            aws.String(bucket),
		Prefix:            aws.String(prefix),
		MaxKeys:           maxKeys,
		ContinuationToken: continuationToken,
	}

	resp, err := c.s3.ListObjectsV2(ctx, input)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list objects in S3: %w", err)
	}

	files := make([]S3File, 0, len(resp.Contents))
	for _, obj := range resp.Contents {
		name := aws.ToString(obj.Key)
		if prefix != "" && strings.HasPrefix(name, prefix) {
			name = strings.TrimPrefix(name, prefix)
		}
		files = append(files, S3File{
			Path:         aws.ToString(obj.Key),
			Name:         name,
			Size:         aws.ToInt64(obj.Size),
			LastModified: obj.LastModified.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return files, resp.NextContinuationToken, nil
}

func (c *Client) ListFilesAll(ctx context.Context, bucket string, prefix string) ([]S3File, error) {
	var allFiles []S3File
	var token *string
	max := int32(1000)

	for {
		files, nextToken, err := c.ListFilesPaged(ctx, bucket, prefix, &max, token)
		if err != nil {
			return nil, err
		}
		allFiles = append(allFiles, files...)
		if nextToken == nil || *nextToken == "" {
			break
		}
		token = nextToken
	}
	return allFiles, nil
}

// compat (una sola pagina)
func (c *Client) ListFiles(ctx context.Context, bucket string, prefix string, maxKeys *int32) ([]S3File, error) {
	files, _, err := c.ListFilesPaged(ctx, bucket, prefix, maxKeys, nil)
	return files, err
}

/* -------------------- WALK (paginato + callback) -------------------- */

func (c *Client) WalkPrefix(
	ctx context.Context,
	bucket string,
	prefix string,
	pageSize int32,
	fn func(obj s3types.Object) error,
) error {
	var token *string

	for {
		input := &s3.ListObjectsV2Input{
			Bucket:            aws.String(bucket),
			Prefix:            aws.String(prefix),
			MaxKeys:           aws.Int32(pageSize),
			ContinuationToken: token,
		}

		resp, err := c.s3.ListObjectsV2(ctx, input)
		if err != nil {
			return fmt.Errorf("list error: %w", err)
		}

		for _, obj := range resp.Contents {
			// escludi placeholder "cartella"
			if obj.Key != nil && !(strings.HasSuffix(aws.ToString(obj.Key), "/") && aws.ToInt64(obj.Size) == 0) {
				if err := fn(obj); err != nil {
					return err
				}
			}
		}

		if resp.NextContinuationToken == nil || *resp.NextContinuationToken == "" {
			break
		}
		token = resp.NextContinuationToken
	}
	return nil
}

/* -------------------- DOWNLOAD / UPLOAD -------------------- */

func (c *Client) DownloadFile(ctx context.Context, bucket, key, localPath string) error {
	out, err := c.s3.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err != nil {
		return fmt.Errorf("failed to get object from S3: %w", err)
	}
	defer out.Body.Close()

	f, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, out.Body); err != nil {
		return fmt.Errorf("failed to write to local file: %w", err)
	}
	return nil
}

func (c *Client) UploadFile(ctx context.Context, bucket, key string, file *os.File) (interface{}, error) {
	const threshold = 100 * 1024 * 1024

	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat error: %w", err)
	}
	size := info.Size()
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("seek error: %w", err)
	}

	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	mime := http.DetectContentType(buf[:n])
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("rewind error: %w", err)
	}

	fmt.Printf("Uploading to s3://%s/%s (%.2fMB, type: %s)\n", bucket, key, float64(size)/(1024*1024), mime)

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
