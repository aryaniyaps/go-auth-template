package captcha

import "errors"

var (
	// ErrEmptyToken is returned when an empty captcha token is provided
	ErrEmptyToken = errors.New("captcha token cannot be empty")

	// ErrInvalidToken is returned when the captcha token is invalid
	ErrInvalidToken = errors.New("invalid captcha token")

	// ErrVerificationFailed is returned when captcha verification fails
	ErrVerificationFailed = errors.New("captcha verification failed")

	// ErrProviderNotConfigured is returned when the captcha provider is not properly configured
	ErrProviderNotConfigured = errors.New("captcha provider not configured")

	// ErrUnsupportedProvider is returned when an unsupported captcha provider is specified
	ErrUnsupportedProvider = errors.New("unsupported captcha provider")
)