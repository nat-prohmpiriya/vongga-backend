package adapter

import (
	"context"
	"fmt"
	"vongga_api/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config" // เพิ่ม alias เพื่อไม่ให้ชนกับ local config
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// NewCloudflareAdapter creates a new S3 client for Cloudflare R2
// NewCloudflareAdapter creates a new S3 client for Cloudflare R2
func NewCloudflareAdapter(cfg *config.Config) (*s3.Client, error) {
	// Create AWS config
	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion("auto"),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				cfg.R2AccessKeyID,
				cfg.R2SecretKey,
				"",
			),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client with custom endpoint
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.R2URLSpecial)
	})

	return client, nil
}
