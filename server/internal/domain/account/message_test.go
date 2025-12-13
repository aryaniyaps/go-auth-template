package account

import (
	"bytes"
	"context"
	"log"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDummyMessageSender_SendSMS(t *testing.T) {
	tests := []struct {
		name        string
		phoneNumber string
		message     string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid SMS message",
			phoneNumber: "+12345678901",
			message:     "Test message",
			expectError: false,
		},
		{
			name:        "invalid phone number",
			phoneNumber: "invalid-phone",
			message:     "Test message",
			expectError: true,
			errorMsg:    "invalid phone number",
		},
		{
			name:        "empty phone number",
			phoneNumber: "",
			message:     "Test message",
			expectError: true,
			errorMsg:    "phone number cannot be empty",
		},
		{
			name:        "message too long",
			phoneNumber: "+12345678901",
			message:     strings.Repeat("x", 1601),
			expectError: true,
			errorMsg:    "message too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create logger with buffer for capturing logs
			var logBuffer bytes.Buffer
			logger := log.New(&logBuffer, "", 0)

			config := &DummySMSConfig{
				Enabled:     true,
				LogMessages: true,
				FromNumber:  "+12345678901",
			}

			sender := NewDummyMessageSender(config, logger)
			err := sender.SendSMS(context.Background(), tt.phoneNumber, tt.message)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
				// Check that message was logged
				logOutput := logBuffer.String()
				assert.Contains(t, logOutput, tt.phoneNumber)
				assert.Contains(t, logOutput, tt.message)
				assert.Contains(t, logOutput, "[SMS DUMMY]")
			}
		})
	}
}

func TestDummyMessageSender_ValidatePhoneNumber(t *testing.T) {
	tests := []struct {
		name        string
		phoneNumber string
		expectError bool
	}{
		{
			name:        "valid US phone number",
			phoneNumber: "+12345678901",
			expectError: false,
		},
		{
			name:        "valid UK phone number",
			phoneNumber: "+447911123456",
			expectError: false,
		},
		{
			name:        "invalid phone number",
			phoneNumber: "1234567890",
			expectError: true,
		},
		{
			name:        "empty phone number",
			phoneNumber: "",
			expectError: true,
		},
		{
			name:        "malformed phone number",
			phoneNumber: "abc-def-ghij",
			expectError: true,
		},
	}

	config := &DummySMSConfig{}
	sender := NewDummyMessageSender(config, nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sender.ValidatePhoneNumber(tt.phoneNumber)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDummyMessageSender_DisabledSMS(t *testing.T) {
	config := &DummySMSConfig{
		Enabled: false,
	}
	sender := NewDummyMessageSender(config, nil)

	err := sender.SendSMS(context.Background(), "+12345678901", "Test message")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SMS sending is disabled")
}

func TestDummyMessageSender_ContextCancellation(t *testing.T) {
	config := &DummySMSConfig{
		Enabled: true,
	}
	sender := NewDummyMessageSender(config, nil)

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := sender.SendSMS(ctx, "+12345678901", "Test message")
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestMessageFormatting(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		expected string
	}{
		{
			name:     "verification message format",
			token:    "123456",
			expected: "Your verification code is: 123456. This code will expire in 24 hours.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message := FormatVerificationMessage(tt.token)
			assert.Equal(t, tt.expected, message)
		})
	}

	// Test account update message
	updateMessage := FormatAccountUpdateMessage("profile")
	assert.Contains(t, updateMessage, "Your account profile has been successfully updated")
	assert.Contains(t, updateMessage, "contact support immediately")
}

func TestValidateMessageContent(t *testing.T) {
	tests := []struct {
		name        string
		message     string
		expectError bool
	}{
		{
			name:        "valid message",
			message:     "This is a valid message",
			expectError: false,
		},
		{
			name:        "empty message",
			message:     "",
			expectError: true,
		},
		{
			name:        "whitespace only message",
			message:     "   \t\n   ",
			expectError: true,
		},
		{
			name:        "too long message",
			message:     strings.Repeat("x", 1601),
			expectError: true,
		},
		{
			name:        "message with script tag",
			message:     "Click <script>alert('xss')</script> here",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMessageContent(tt.message)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}