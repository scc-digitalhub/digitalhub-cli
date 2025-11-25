// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package transfer

import (
	"context"
	s3client "dhcli/configs"
	"dhcli/sdk/config"
	"fmt"
)

type TransferService struct {
	http config.CoreHTTP
	s3   *s3client.Client
}

func NewTransferService(ctx context.Context, conf config.Config) (*TransferService, error) {
	httpc := config.NewHTTPCore(nil, conf.Core)

	s3c, err := s3client.NewClient(ctx, s3client.Config{
		AccessKey:   conf.S3.AccessKey,
		SecretKey:   conf.S3.SecretKey,
		AccessToken: conf.S3.SessionToken,
		Region:      conf.S3.Region,
		EndpointURL: conf.S3.EndpointURL,
	})
	if err != nil {
		return nil, fmt.Errorf("S3 init failed: %w", err)
	}

	return &TransferService{http: httpc, s3: s3c}, nil
}
