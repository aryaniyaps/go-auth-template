package account

import (
	"errors"
	"fmt"
)

// Well-defined error types for account domain operations
// These errors can be pattern matched using errors.Is() and errors.As()

// Base error types
var (
	// Repository errors
	ErrAccountNotFound    = errors.New("account not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrPhoneAlreadyExists = errors.New("phone number already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrTokenExpired       = errors.New("token has expired")
	ErrTokenNotFound      = errors.New("token not found")

	// Validation errors
	ErrInvalidInput        = errors.New("invalid input")
	ErrInvalidEmail        = errors.New("invalid email format")
	ErrInvalidPhoneNumber  = errors.New("invalid phone number format")
	ErrInvalidFullName     = errors.New("invalid full name")
	ErrInvalidFile         = errors.New("invalid file")
	ErrFileTooLarge        = errors.New("file too large")
	ErrUnsupportedFileType = errors.New("unsupported file type")

	// Business logic errors
	ErrAccountDisabled    = errors.New("account is disabled")
	ErrVerificationFailed = errors.New("verification failed")
	ErrSMSSendingFailed   = errors.New("SMS sending failed")
	ErrEmailSendingFailed = errors.New("email sending failed")

	// Configuration errors
	ErrS3NotConfigured = errors.New("S3 client not configured")
	ErrSMSServiceDown  = errors.New("SMS service unavailable")
)

// Detailed error types with additional context
type ValidationError struct {
	Field   string
	Message string
	Err     error
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation failed for field '%s': %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation failed: %s", e.Message)
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

type ServiceError struct {
	Operation string
	Message   string
	Err       error
}

func (e *ServiceError) Error() string {
	return fmt.Sprintf("service error during %s: %s", e.Operation, e.Message)
}

func (e *ServiceError) Unwrap() error {
	return e.Err
}

type RepositoryError struct {
	Operation string
	Entity    string
	Message   string
	Err       error
}

func (e *RepositoryError) Error() string {
	return fmt.Sprintf("repository error during %s on %s: %s", e.Operation, e.Entity, e.Message)
}

func (e *RepositoryError) Unwrap() error {
	return e.Err
}

// Helper functions to create detailed errors
func NewValidationError(field, message string, err error) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
		Err:     err,
	}
}

func NewServiceError(operation, message string, err error) *ServiceError {
	return &ServiceError{
		Operation: operation,
		Message:   message,
		Err:       err,
	}
}

func NewRepositoryError(operation, entity, message string, err error) *RepositoryError {
	return &RepositoryError{
		Operation: operation,
		Entity:    entity,
		Message:   message,
		Err:       err,
	}
}

// Constants for error messages
const (
	MsgAccountNotFound           = "account not found"
	MsgInvalidEmailFormat        = "email format is invalid"
	MsgInvalidPhoneNumberFormat  = "phone number format is invalid"
	MsgFullNameRequired          = "full name is required and cannot be empty"
	MsgTermsVersionRequired      = "terms version cannot be empty"
	MsgInvalidAnalyticsPreference = "analytics preference must be 'enabled', 'disabled', or 'undecided'"
	MsgFileRequired              = "file is required"
	MsgFilenameRequired          = "filename is required"
	MsgFileEmpty                 = "file is empty"
	MsgAvatarFileSizeExceeded    = "avatar file size exceeds maximum allowed size"
	MsgPhoneNumberRequired       = "phone number is required"
	MsgVerificationTokenRequired = "verification token is required"
	MsgVerificationTokenExpired  = "verification token has expired"
	MsgVerificationTokenInvalid  = "verification token is invalid"
	MsgSMSServiceDisabled        = "SMS service is currently disabled"
	MsgMessageTooLong            = "message exceeds maximum allowed length"
	MsgMessageRequired           = "message is required"
	MsgMessageUnsafe             = "message contains potentially unsafe content"
	MsgS3ClientRequired          = "S3 client is required for this operation"
)