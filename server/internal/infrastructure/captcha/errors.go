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

	// ErrMissingSecretKey is returned when the secret key is not configured
	ErrMissingSecretKey = errors.New("captcha secret key is required")

	// ErrInvalidSecretKey is returned when the secret key is invalid
	ErrInvalidSecretKey = errors.New("invalid captcha secret key")

	// ErrBadRequest is returned when the request is malformed
	ErrBadRequest = errors.New("bad request to captcha provider")

	// ErrExpiredToken is returned when the captcha token has expired
	ErrExpiredToken = errors.New("captcha token has expired")
)