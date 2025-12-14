package auth

import (
	"errors"
	"fmt"
)

// Well-defined error types for auth domain operations
// These errors can be pattern matched using errors.Is() and errors.As()

// Base error types
var (
	// Authentication errors
	ErrInvalidCredentials      = errors.New("invalid credentials")
	ErrAccountNotFound         = errors.New("account not found")
	ErrAccountDisabled         = errors.New("account is disabled")
	ErrAuthenticationFailed    = errors.New("authentication failed")
	ErrTwoFactorRequired       = errors.New("two-factor authentication required")
	ErrInvalidTwoFactorCode    = errors.New("invalid two-factor authentication code")

	// Token and session errors
	ErrSessionNotFound         = errors.New("session not found")
	ErrTokenExpired            = errors.New("token has expired")
	ErrTokenNotFound           = errors.New("token not found")
	ErrInvalidToken            = errors.New("invalid token")
	ErrInvalidOrExpiredToken   = errors.New("token is invalid or expired")
	ErrPasswordResetTokenNotFound = errors.New("password reset token not found")

	// Password errors
	ErrPasswordTooWeak         = errors.New("password is too weak")
	ErrPasswordIncorrect       = errors.New("password is incorrect")
	ErrPasswordResetRequired   = errors.New("password reset required")

	// Email and phone errors
	ErrEmailAlreadyExists      = errors.New("email already exists")
	ErrPhoneAlreadyExists      = errors.New("phone number already exists")
	ErrEmailNotVerified        = errors.New("email not verified")
	ErrInvalidEmailDomain      = errors.New("invalid or disposable email domain")
	ErrEmailAlreadyVerified    = errors.New("email is already verified")

	// WebAuthn errors
	ErrWebAuthnCredentialNotFound = errors.New("webauthn credential not found")
	ErrChallengeNotFound         = errors.New("challenge not found")
	ErrInvalidWebAuthnResponse   = errors.New("invalid webauthn response")

	// OAuth errors
	ErrOAuthCredentialAlreadyExists = errors.New("oauth credential already exists")
	ErrOAuthTokenInvalid           = errors.New("oauth token is invalid")
	ErrOAuthProviderUnsupported    = errors.New("oauth provider not supported")

	// Rate limiting errors
	ErrRateLimitExceeded        = errors.New("rate limit exceeded")
	ErrAccountLocked           = errors.New("account is temporarily locked")
	ErrEmailCooldown           = errors.New("email verification request too frequent")

	// 2FA and recovery errors
	ErrRecoveryCodeInvalid     = errors.New("recovery code is invalid")
	ErrTwoFactorAuthenticationNotFound = errors.New("two factor authentication challenge not found")
	ErrTemporaryTwoFactorNotFound      = errors.New("temporary two factor challenge not found")
	ErrTwoFactorNotEnabled     = errors.New("two-factor authentication not enabled")

	// CAPTCHA errors
	ErrCAPTCHARequired         = errors.New("captcha verification required")
	ErrCAPTCHAInvalid          = errors.New("captcha verification failed")
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
	MsgInvalidCredentials         = "invalid email or password"
	MsgAccountNotFound            = "account not found"
	MsgAccountDisabled            = "account is disabled"
	MsgTwoFactorRequired          = "two-factor authentication is required"
	MsgInvalidTwoFactorCode       = "invalid two-factor authentication code"
	MsgPasswordTooWeak            = "password must be at least 8 characters and contain uppercase, lowercase, digit, and special character"
	MsgPasswordIncorrect          = "current password is incorrect"
	MsgEmailAlreadyExists         = "email is already registered"
	MsgPhoneAlreadyExists         = "phone number is already registered"
	MsgEmailNotVerified           = "email address must be verified before this operation"
	MsgInvalidEmailDomain         = "email domain is not allowed or is disposable"
	MsgRateLimitExceeded          = "too many requests, please try again later"
	MsgAccountLocked              = "account is temporarily locked due to too many failed attempts"
	MsgEmailCooldown              = "please wait before requesting another verification email"
	MsgCAPTCHARequired            = "captcha verification is required for this operation"
	MsgCAPTCHAInvalid             = "captcha verification failed"
	MsgRecoveryCodeInvalid        = "recovery code is invalid or has been used"
	MsgTwoFactorNotEnabled        = "two-factor authentication is not enabled for this account"
	MsgOAuthTokenInvalid          = "oauth token is invalid or expired"
	MsgOAuthProviderUnsupported   = "oauth provider is not supported"
	MsgInvalidWebAuthnResponse    = "webauthn response is invalid"
	MsgChallengeNotFound          = "challenge not found or expired"
	MsgInvalidToken               = "token is invalid"
)