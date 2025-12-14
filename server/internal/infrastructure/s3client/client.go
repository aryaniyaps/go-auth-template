package s3client

import (
	"context"
	"fmt"
	appconfig "server/internal/config"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// NewS3Client creates a new S3 client with default configuration
func NewS3Client(ctx context.Context, region string) (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}
	return s3.NewFromConfig(cfg), nil
}

// NewS3ClientFromEnv creates an S3 client using environment variables
func NewS3ClientFromEnv(ctx context.Context) (*s3.Client, error) {
	region := "us-east-1" // Default region
	if envRegion := ctx.Value("aws_region"); envRegion != nil {
		if r, ok := envRegion.(string); ok {
			region = r
		}
	}
	return NewS3Client(ctx, region)
}

// NewS3ClientProvider creates an S3 client using configuration for FX
func NewS3ClientProvider(cfg *appconfig.Config) (*s3.Client, error) {
	// Return nil client gracefully if S3 is not configured
	// This allows the service to work in development without S3
	if cfg.S3Bucket == "" {
		return nil, nil
	}

	client, err := NewS3Client(context.Background(), cfg.S3Region)
	if err != nil {
		return nil, err
	}

	return client, nil
}
