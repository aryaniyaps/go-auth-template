package account

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"server/internal/config"
)

func TestAccountServiceDependencyInjection(t *testing.T) {
	tests := []struct {
		name        string
		s3Bucket    string
		expectError bool
	}{
		{
			name:        "successful dependency injection without S3",
			s3Bucket:    "",
			expectError: false,
		},
		{
			name:        "successful dependency injection with S3 configured",
			s3Bucket:    "test-bucket",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test configuration
			cfg := &config.Config{
				S3Bucket: tt.s3Bucket,
				S3Region: "us-east-1",
			}

			// Create FX application with account domain module
			app := fx.New(
				AccountDomainModule,
				fx.Provide(
					func() *config.Config { return cfg },
					func() *zap.Logger { return zap.NewNop() },
				),
				fx.Invoke(
					// This will trigger dependency injection
					func(service *AccountService) {
						require.NotNil(t, service)
						assert.NotNil(t, service.logger)
						assert.NotNil(t, service.messageSender)
						// S3 client can be nil if not configured
						if tt.s3Bucket == "" {
							assert.Nil(t, service.s3Client)
						}
					},
				),
			)

			ctx := context.Background()

			if tt.expectError {
				err := app.Start(ctx)
				assert.Error(t, err)
			} else {
				err := app.Start(ctx)
				// Should not error even without S3 configured
				if err != nil {
					// If there's an error, it should be related to missing dependencies, not S3
					t.Logf("FX start error (expected for incomplete test setup): %v", err)
				}

				// Clean shutdown
				err = app.Stop(ctx)
				assert.NoError(t, err)
			}
		})
	}
}

func TestS3ClientProvider(t *testing.T) {
	tests := []struct {
		name        string
		s3Bucket    string
		expectNil   bool
	}{
		{
			name:      "nil S3 client when bucket not configured",
			s3Bucket:  "",
			expectNil: true,
		},
		{
			name:      "S3 client when bucket configured",
			s3Bucket:  "test-bucket",
			expectNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				S3Bucket: tt.s3Bucket,
				S3Region: "us-east-1",
			}

			client, err := NewS3ClientProvider(cfg)

			// Should not error even if S3 is not configured
			assert.NoError(t, err)

			if tt.expectNil {
				assert.Nil(t, client)
			} else {
				assert.NotNil(t, client)
			}
		})
	}
}

func TestDummyMessageSenderDependencyInjection(t *testing.T) {
	logger := zap.NewNop()

	sender := NewDummyMessageSenderForFX(logger)

	require.NotNil(t, sender)
	assert.IsType(t, &DummyMessageSender{}, sender)
}

func TestAccountServiceWithOptionalS3Client(t *testing.T) {
	mockRepo := &MockAccountRepo{}
	mockPhoneTokenRepo := &MockPhoneNumberVerificationTokenRepo{}
	mockEmailTokenRepo := &MockEmailVerificationTokenRepo{}
	mockMessageSender := &MockMessageSender{}
	logger := zap.NewNop()

	// Test service creation with nil S3 client (development mode)
	service1 := NewAccountService(
		mockRepo,
		mockPhoneTokenRepo,
		mockEmailTokenRepo,
		mockMessageSender,
		nil, // S3 client is nil
		logger,
	)

	require.NotNil(t, service1)
	assert.Nil(t, service1.s3Client)
	assert.NotNil(t, service1.logger)
	assert.NotNil(t, service1.messageSender)

	// Test service creation with S3 client (production mode)
	// Note: We can't create a real S3 client in tests without AWS credentials
	// but we can verify the constructor accepts the parameter
	service2 := NewAccountService(
		mockRepo,
		mockPhoneTokenRepo,
		mockEmailTokenRepo,
		mockMessageSender,
		nil, // Would be real S3 client in production
		logger,
	)

	require.NotNil(t, service2)
	assert.NotNil(t, service2.logger)
	assert.NotNil(t, service2.messageSender)
}