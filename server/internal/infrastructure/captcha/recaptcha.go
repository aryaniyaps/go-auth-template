package captcha

import (
	"context"
	"fmt"
)

// ReCaptchaVerifier implements captcha verification using Google reCAPTCHA
type ReCaptchaVerifier struct {
	secretKey string
}

// NewReCaptchaVerifier creates a new reCAPTCHA verifier
func NewReCaptchaVerifier(config *CaptchaConfig) (CaptchaVerifier, error) {
	if config.ReCaptchaSecretKey == "" {
		return nil, fmt.Errorf("%w: reCAPTCHA secret key is required", ErrProviderNotConfigured)
	}

	// TODO: Implement reCAPTCHA verification
	return &ReCaptchaVerifier{
		secretKey: config.ReCaptchaSecretKey,
	}, nil
}

// VerifyToken verifies a reCAPTCHA token
func (r *ReCaptchaVerifier) VerifyToken(ctx context.Context, token string) (bool, error) {
	// TODO: Implement actual reCAPTCHA verification
	return false, fmt.Errorf("reCAPTCHA verifier not yet implemented")
}