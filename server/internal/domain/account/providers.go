package account

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"server/internal/config"
)

// AccountDomainModule contains all account domain repositories and services for dependency injection
var AccountDomainModule = fx.Options(
	fx.Provide(
		NewAccountRepo,
		NewEmailVerificationTokenRepo,
		NewPhoneNumberVerificationTokenRepo,
		NewAccountService,
		NewDummyMessageSenderForFX,
		NewS3ClientProvider,
	),
)

// NewDummyMessageSenderForFX creates a new dummy message sender for FX dependency injection
func NewDummyMessageSenderForFX(logger *zap.Logger) MessageSender {
	// Convert zap logger to standard logger for compatibility with existing code
	stdLogger := log.Default()

	config := &DummySMSConfig{
		Enabled:      true,
		LogMessages:  true,
		FromNumber:   "+12345678901",
		ValidateOnly: false,
	}

	return NewDummyMessageSender(config, stdLogger)
}

// NewS3ClientProvider creates an S3 client using configuration for FX
func NewS3ClientProvider(cfg *config.Config) (*s3.Client, error) {
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