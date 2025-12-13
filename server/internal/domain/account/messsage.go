package account

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/nyaruka/phonenumbers"
)

// MessageSender interface following BaseMessageSender pattern from Python
type MessageSender interface {
	SendSMS(ctx context.Context, phoneNumber, message string) error
	ValidatePhoneNumber(phoneNumber string) error
}

// DummyMessageSender struct that logs messages instead of sending
type DummyMessageSender struct {
	config *DummySMSConfig
	logger *log.Logger
}

type DummySMSConfig struct {
	Enabled      bool
	LogMessages  bool
	FromNumber   string
	ValidateOnly bool
}

// NewDummyMessageSender creates a new dummy SMS sender
func NewDummyMessageSender(config *DummySMSConfig, logger *log.Logger) MessageSender {
	if config == nil {
		config = &DummySMSConfig{
			Enabled:     true,
			LogMessages: true,
			FromNumber:  "+12345678901",
		}
	}

	if logger == nil {
		logger = log.Default()
	}

	return &DummyMessageSender{
		config: config,
		logger: logger,
	}
}

// SendSMS implements MessageSender interface for dummy provider
func (d *DummyMessageSender) SendSMS(ctx context.Context, phoneNumber, message string) error {
	if err := d.ValidatePhoneNumber(phoneNumber); err != nil {
		return fmt.Errorf("invalid phone number: %w", err)
	}

	if !d.config.Enabled {
		return fmt.Errorf("SMS sending is disabled")
	}

	// Validate message length
	if len(message) > 1600 {
		return fmt.Errorf("message too long: max 1600 characters")
	}

	// Log the message instead of sending
	if d.config.LogMessages {
		d.logger.Printf("[SMS DUMMY] To: %s | From: %s | Message: %s",
			phoneNumber, d.config.FromNumber, message)
	}

	// Simulate network delay
	select {
	case <-time.After(50 * time.Millisecond):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// ValidatePhoneNumber validates phone number format
func (d *DummyMessageSender) ValidatePhoneNumber(phoneNumber string) error {
	if strings.TrimSpace(phoneNumber) == "" {
		return fmt.Errorf("phone number cannot be empty")
	}

	// Parse and validate phone number using phonenumbers library
	parsedNum, err := phonenumbers.Parse(phoneNumber, "")
	if err != nil {
		return fmt.Errorf("failed to parse phone number: %w", err)
	}

	if !phonenumbers.IsValidNumber(parsedNum) {
		return fmt.Errorf("invalid phone number format")
	}

	return nil
}

// Message formatting utilities
func FormatVerificationMessage(token string) string {
	return fmt.Sprintf("Your verification code is: %s. This code will expire in 24 hours.", token)
}

func FormatAccountUpdateMessage(updateType string) string {
	return fmt.Sprintf("Your account %s has been successfully updated. If you did not make this change, please contact support immediately.", updateType)
}

// Validate message content
func ValidateMessageContent(message string) error {
	if strings.TrimSpace(message) == "" {
		return fmt.Errorf("message cannot be empty")
	}

	if len(message) > 1600 {
		return fmt.Errorf("message too long: max 1600 characters")
	}

	// Basic content validation
	if strings.Contains(message, "<script>") {
		return fmt.Errorf("message contains potentially unsafe content")
	}

	return nil
}