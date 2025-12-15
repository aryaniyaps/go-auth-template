package captcha

import (
	"context"
	"fmt"
)

// HCaptchaVerifier implements captcha verification using hCaptcha
type HCaptchaVerifier struct {
	secretKey string
}

// NewHCaptchaVerifier creates a new hCaptcha verifier
func NewHCaptchaVerifier(config *CaptchaConfig) (CaptchaVerifier, error) {
	if config.HCaptchaSecretKey == "" {
		return nil, fmt.Errorf("%w: hCaptcha secret key is required", ErrProviderNotConfigured)
	}

	// TODO: Implement hCaptcha verification
	return &HCaptchaVerifier{
		secretKey: config.HCaptchaSecretKey,
	}, nil
}

// VerifyToken verifies an hCaptcha token
func (h *HCaptchaVerifier) VerifyToken(ctx context.Context, token string) (bool, error) {
	// TODO: Implement actual hCaptcha verification
	return false, fmt.Errorf("hCaptcha verifier not yet implemented")
}
