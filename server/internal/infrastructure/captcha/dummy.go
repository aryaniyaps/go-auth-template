package captcha

import (
	"context"
	"fmt"
)

// DummyVerifier provides a mock captcha verification implementation
// useful for development and testing environments.
type DummyVerifier struct {
	alwaysSucceed bool
	forceFail     bool
}

// NewDummyVerifier creates a new dummy captcha verifier.
// By default, it always succeeds to make development easier.
func NewDummyVerifier() *DummyVerifier {
	return &DummyVerifier{
		alwaysSucceed: true,
		forceFail:     false,
	}
}

// NewDummyVerifierWithConfig creates a dummy verifier with specific behavior.
// Use this for testing different scenarios.
func NewDummyVerifierWithConfig(alwaysSucceed bool) *DummyVerifier {
	return &DummyVerifier{
		alwaysSucceed: alwaysSucceed,
		forceFail:     false,
	}
}

// VerifyToken implements BaseCaptchaVerifier interface.
// For the dummy verifier, it either always succeeds or always fails
// based on the configuration. This is useful for testing different
// scenarios without needing real captcha tokens.
func (dv *DummyVerifier) VerifyToken(ctx context.Context, captchaToken string) (bool, error) {
	// If force fail is set, always return failure
	if dv.forceFail {
		return false, ErrInvalidToken
	}

	// If always succeed is true, always return success
	if dv.alwaysSucceed {
		return true, nil
	}

	// Default behavior: validate token format (basic validation)
	if captchaToken == "" {
		return false, ErrEmptyToken
	}

	// For testing purposes, accept any non-empty token
	return true, nil
}

// SetForceFail configures the verifier to always fail.
// Useful for testing error handling paths.
func (dv *DummyVerifier) SetForceFail(forceFail bool) {
	dv.forceFail = forceFail
}

// SetAlwaysSucceed configures whether the verifier should always succeed.
func (dv *DummyVerifier) SetAlwaysSucceed(alwaysSucceed bool) {
	dv.alwaysSucceed = alwaysSucceed
}

// HealthCheck verifies the dummy verifier is working.
// Since this is a mock implementation, it always succeeds.
func (dv *DummyVerifier) HealthCheck(ctx context.Context) error {
	return nil
}

// String returns a string representation of the verifier.
func (dv *DummyVerifier) String() string {
	return fmt.Sprintf("DummyVerifier{alwaysSucceed: %v, forceFail: %v}", dv.alwaysSucceed, dv.forceFail)
}