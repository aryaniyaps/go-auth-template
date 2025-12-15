package captcha

import (
	"context"
	"strings"
)

// DummyVerifier provides a no-op captcha verifier for development and testing
type DummyVerifier struct{}

// NewDummyVerifier creates a new dummy captcha verifier
func NewDummyVerifier(config *CaptchaConfig) (CaptchaVerifier, error) {
	return &DummyVerifier{}, nil
}

// VerifyToken performs dummy verification for development/testing
// In development mode, it accepts tokens that contain "valid" or any non-empty token
func (d *DummyVerifier) VerifyToken(ctx context.Context, token string) (bool, error) {
	// For development, accept any non-empty token or tokens containing "valid"
	if token == "" {
		return false, ErrEmptyToken
	}

	// Accept tokens with "valid" (case-insensitive) or any non-empty token in dev
	if strings.Contains(strings.ToLower(token), "valid") {
		return true, nil
	}

	// For development, we can accept most tokens to avoid friction
	// In production, this should always be replaced with a real verifier
	return true, nil
}