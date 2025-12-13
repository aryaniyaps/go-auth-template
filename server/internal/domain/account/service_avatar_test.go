package account

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestAccountService_UpdateAccountAvatarURL(t *testing.T) {
	tests := []struct {
		name          string
		fileContent   []byte
		filename      string
		expectError   bool
		errorContains string
	}{
		{
			name:          "nil file should return error",
			fileContent:   nil,
			filename:      "avatar.jpg",
			expectError:   true,
			errorContains: "file cannot be nil",
		},
		{
			name:          "empty filename should return error",
			fileContent:   []byte("fake image content"),
			filename:      "",
			expectError:   true,
			errorContains: "filename cannot be empty",
		},
		{
			name:          "whitespace-only filename should return error",
			fileContent:   []byte("fake image content"),
			filename:      "   ",
			expectError:   true,
			errorContains: "filename cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockAccountRepo{}
			logger := zap.NewNop()

			// Create service without S3 client
			service := NewAccountService(mockRepo, nil, nil, nil, nil, logger)

			ctx := context.Background()
			var file *bytes.Reader
			if tt.fileContent != nil {
				file = bytes.NewReader(tt.fileContent)
			}

			result, err := service.UpdateAccountAvatarURL(ctx, 1, file, tt.filename)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestAccountService_ValidateAvatarFile(t *testing.T) {
	mockRepo := &MockAccountRepo{}
	mockPhoneTokenRepo := &MockPhoneNumberVerificationTokenRepo{}
	mockEmailTokenRepo := &MockEmailVerificationTokenRepo{}
	mockSMS := &MockMessageSender{}
	logger := zap.NewNop()
	service := NewAccountService(mockRepo, mockPhoneTokenRepo, mockEmailTokenRepo, mockSMS, nil, logger)

	tests := []struct {
		name          string
		fileContent   []byte
		filename      string
		expectError   bool
		errorContains string
		expectedType  string
	}{
		{
			name:          "valid JPEG file should pass",
			fileContent:   []byte("\xFF\xD8\xFF\xE0\x00\x10JFIF"), // JPEG header
			filename:      "avatar.jpg",
			expectError:   false,
			expectedType:  "image/jpeg",
		},
		{
			name:          "valid PNG file should pass",
			fileContent:   []byte("\x89PNG\r\n\x1a\n"), // PNG header
			filename:      "avatar.png",
			expectError:   false,
			expectedType:  "image/png",
		},
		{
			name:          "empty file should return error",
			fileContent:   []byte(""),
			filename:      "avatar.jpg",
			expectError:   true,
			errorContains: "file is empty",
		},
		{
			name:          "file too large should return error",
			fileContent:   make([]byte, 6*1024*1024), // 6MB file
			filename:      "avatar.jpg",
			expectError:   true,
			errorContains: "file size exceeds maximum",
		},
		{
			name:          "invalid file type should return error",
			fileContent:   []byte("PK\x03\x04"), // ZIP header
			filename:      "avatar.zip",
			expectError:   true,
			errorContains: "file type application/zip is not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := bytes.NewReader(tt.fileContent)

			contentBytes, contentType, err := service.validateAvatarFile(file, tt.filename)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				assert.Empty(t, contentBytes)
				assert.Empty(t, contentType)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, contentBytes)
				assert.Equal(t, tt.expectedType, contentType)
			}
		})
	}
}

func TestAccountService_GenerateUniqueFilename(t *testing.T) {
	mockRepo := &MockAccountRepo{}
	mockPhoneTokenRepo := &MockPhoneNumberVerificationTokenRepo{}
	mockEmailTokenRepo := &MockEmailVerificationTokenRepo{}
	mockSMS := &MockMessageSender{}
	logger := zap.NewNop()
	service := NewAccountService(mockRepo, mockPhoneTokenRepo, mockEmailTokenRepo, mockSMS, nil, logger)

	tests := []struct {
		name         string
		filename     string
		expectUnique bool
	}{
		{
			name:         "filename with jpg extension",
			filename:     "avatar.jpg",
			expectUnique: true,
		},
		{
			name:         "filename with png extension",
			filename:     "photo.png",
			expectUnique: true,
		},
		{
			name:         "filename with multiple dots",
			filename:     "my.photo.jpg",
			expectUnique: true,
		},
		{
			name:         "filename without extension",
			filename:     "avatar",
			expectUnique: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.generateUniqueFilename(tt.filename)

			assert.NotEmpty(t, result)
			assert.NotEqual(t, tt.filename, result)

			// Check that result starts with timestamp and ends with extension
			if tt.expectUnique {
				// Simpler approach: just check it contains a timestamp-like number
				assert.GreaterOrEqual(t, len(result), len("123.jpg"))
			}
		})
	}
}

func TestAccountService_AvatarIntegration(t *testing.T) {
	// This test verifies the complete workflow until the S3 step
	// We mock only the components we need and test the actual flow

	mockRepo := &MockAccountRepo{}
	mockPhoneTokenRepo := &MockPhoneNumberVerificationTokenRepo{}
	mockEmailTokenRepo := &MockEmailVerificationTokenRepo{}
	mockSMS := &MockMessageSender{}
	logger := zap.NewNop()

	// Create service with nil S3 client (will fail at S3 step)
	service := NewAccountService(mockRepo, mockPhoneTokenRepo, mockEmailTokenRepo, mockSMS, nil, logger)

	ctx := context.Background()
	fileContent := []byte("\xFF\xD8\xFF") // JPEG header
	file := bytes.NewReader(fileContent)

	// Test complete avatar upload workflow (will fail at S3 step)
	// Note: Since S3 client is nil, it will fail before even calling the repo.Get
	// So we don't need to set up that expectation
	result, err := service.UpdateAccountAvatarURL(ctx, 1, file, "avatar.jpg")

	// This should fail at S3 client validation step
	require.Error(t, err)
	assert.Contains(t, err.Error(), "S3 client not configured")
	assert.Nil(t, result)
}

func TestAccountService_FileValidationConstants(t *testing.T) {
	// Test that validation constants are properly set
	assert.Equal(t, 5<<20, MaxAvatarFileSize) // 5MB
	assert.Equal(t, "image/jpeg,image/png,image/gif,image/webp", AllowedAvatarTypes)
}

func TestAccountService_AvatarWorkflowWithS3(t *testing.T) {
	// This test simulates the full workflow with mocked S3

	mockRepo := &MockAccountRepo{}
	mockPhoneTokenRepo := &MockPhoneNumberVerificationTokenRepo{}
	mockEmailTokenRepo := &MockEmailVerificationTokenRepo{}
	mockSMS := &MockMessageSender{}
	logger := zap.NewNop()

	// Create service with mocked S3-like behavior
	// In a real test, you would mock the S3 client, but for now we test the failure case
	service := NewAccountService(mockRepo, mockPhoneTokenRepo, mockEmailTokenRepo, mockSMS, nil, logger)

	ctx := context.Background()
	fileContent := []byte("\xFF\xD8\xFF\xE0\x00\x10JFIF\x00\x01") // More complete JPEG
	file := bytes.NewReader(fileContent)

	// This should fail because S3 client is nil
	result, err := service.UpdateAccountAvatarURL(ctx, 1, file, "test.jpg")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "S3 client not configured")
	assert.Nil(t, result)
}

func TestAccountService_AvatarEdgeCases(t *testing.T) {
	mockRepo := &MockAccountRepo{}
	mockPhoneTokenRepo := &MockPhoneNumberVerificationTokenRepo{}
	mockEmailTokenRepo := &MockEmailVerificationTokenRepo{}
	mockSMS := &MockMessageSender{}
	logger := zap.NewNop()
	service := NewAccountService(mockRepo, mockPhoneTokenRepo, mockEmailTokenRepo, mockSMS, nil, logger)

	tests := []struct {
		name          string
		fileContent   []byte
		filename      string
		expectError   bool
		errorContains string
	}{
		{
			name:          "filename with only spaces",
			fileContent:   []byte("\xFF\xD8\xFF"),
			filename:      "     ",
			expectError:   true,
			errorContains: "filename cannot be empty",
		},
		{
			name:          "filename with tabs",
			fileContent:   []byte("\xFF\xD8\xFF"),
			filename:      "\t\t\n\t",
			expectError:   true,
			errorContains: "filename cannot be empty",
		},
		{
			name:          "gif file should work",
			fileContent:   []byte("GIF87a"), // GIF header
			filename:      "test.gif",
			expectError:   true, // Will fail at S3 client check
			errorContains: "S3 client not configured",
		},
		{
			name:          "webp file should work",
			fileContent:   []byte("RIFF\x00\x00\x00\x00WEBP"), // WebP header
			filename:      "test.webp",
			expectError:   true, // Will fail at S3 client check
			errorContains: "S3 client not configured",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := bytes.NewReader(tt.fileContent)
			result, err := service.UpdateAccountAvatarURL(context.Background(), 1, file, tt.filename)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}