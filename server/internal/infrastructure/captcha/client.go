package captcha

import (
	"context"

	appconfig "server/internal/config"
)

// CaptchaVerifier interface defines the contract for captcha verification
type CaptchaVerifier interface {
	VerifyToken(ctx context.Context, token string) (bool, error)
}

// CaptchaConfig holds configuration for captcha providers
type CaptchaConfig struct {
	Provider           string `mapstructure:"CAPTCHA_PROVIDER"`
	CloudflareSecretKey string `mapstructure:"CLOUDFLARE_SECRET_KEY"`
	CloudflareSiteKey   string `mapstructure:"CLOUDFLARE_SITEKEY"`
	HCaptchaSecretKey  string `mapstructure:"HCAPTCHA_SECRET_KEY"`
	HCaptchaSiteKey    string `mapstructure:"HCAPTCHA_SITEKEY"`
	ReCaptchaSecretKey string `mapstructure:"RECAPTCHA_SECRET_KEY"`
	ReCaptchaSiteKey   string `mapstructure:"RECAPTCHA_SITEKEY"`
}

// CaptchaClient is the main client that handles captcha verification
type CaptchaClient struct {
	config   *CaptchaConfig
	verifier CaptchaVerifier
}

// NewCaptchaClient creates a new captcha client with the appropriate provider
func NewCaptchaClient(config *appconfig.Config) (*CaptchaClient, error) {
	captchaConfig := &CaptchaConfig{
		Provider:            config.CaptchaProvider,
		CloudflareSecretKey: config.CloudflareSecretKey,
		CloudflareSiteKey:   config.CloudflareSiteKey,
		HCaptchaSecretKey:   config.HCaptchaSecretKey,
		HCaptchaSiteKey:     config.HCaptchaSiteKey,
		ReCaptchaSecretKey:  config.ReCaptchaSecretKey,
		ReCaptchaSiteKey:    config.ReCaptchaSiteKey,
	}

	verifier, err := NewCaptchaVerifier(captchaConfig)
	if err != nil {
		return nil, err
	}

	return &CaptchaClient{
		config:   captchaConfig,
		verifier: verifier,
	}, nil
}

// VerifyToken verifies a captcha token using the configured provider
func (c *CaptchaClient) VerifyToken(ctx context.Context, token string) (bool, error) {
	if token == "" {
		return false, ErrEmptyToken
	}

	return c.verifier.VerifyToken(ctx, token)
}

// NewCaptchaVerifier creates the appropriate captcha verifier based on the configuration
func NewCaptchaVerifier(config *CaptchaConfig) (CaptchaVerifier, error) {
	switch config.Provider {
	case "turnstile":
		return NewTurnstileVerifier(config)
	case "recaptcha":
		return NewReCaptchaVerifier(config)
	case "hcaptcha":
		return NewHCaptchaVerifier(config)
	case "dummy":
		return NewDummyVerifier(config)
	default:
		return NewDummyVerifier(config)
	}
}

// NewCaptchaClientProvider creates a captcha verifier for dependency injection
func NewCaptchaClientProvider(config *appconfig.Config) (CaptchaVerifier, error) {
	captchaConfig := &CaptchaConfig{
		Provider:            config.CaptchaProvider,
		CloudflareSecretKey: config.CloudflareSecretKey,
		CloudflareSiteKey:   config.CloudflareSiteKey,
		HCaptchaSecretKey:   config.HCaptchaSecretKey,
		HCaptchaSiteKey:     config.HCaptchaSiteKey,
		ReCaptchaSecretKey:  config.ReCaptchaSecretKey,
		ReCaptchaSiteKey:    config.ReCaptchaSiteKey,
	}

	return NewCaptchaVerifier(captchaConfig)
}