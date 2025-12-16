package captcha

import (
	"go.uber.org/fx"

	appconfig "server/internal/config"
)

const (
	// ProviderDummy represents the dummy captcha provider
	ProviderDummy = "dummy"

	// ProviderTurnstile represents the Cloudflare Turnstile provider
	ProviderTurnstile = "turnstile"
)

// ProviderModule provides captcha verification services using dependency injection
// It follows the established patterns from the email package for consistency.
var ProviderModule = fx.Module("captcha",
	fx.Provide(NewCaptchaVerifierProvider),
)

// NewCaptchaVerifierProvider creates a captcha verifier instance based on the configuration.
// It supports different captcha providers and gracefully falls back to dummy provider
// when the requested provider is not configured or unsupported.
//
// Provider selection logic:
// - "turnstile" → Cloudflare Turnstile (requires CloudflareSecretKey and CloudflareSiteKey)
// - "dummy" or empty → Dummy verifier (default, always works)
// - Other values → Error
//
// Default provider is "dummy" for development environments.
func NewCaptchaVerifierProvider(cfg *appconfig.Config) (BaseCaptchaVerifier, error) {
	if cfg == nil {
		// If no config is provided, default to dummy verifier
		return NewDummyVerifier(), nil
	}

	provider := cfg.CaptchaProvider
	if provider == "" {
		provider = ProviderDummy // Default to dummy if not specified
	}

	switch provider {
	case ProviderDummy:
		return NewDummyVerifier(), nil

	case ProviderTurnstile:
		// Validate required configuration for Turnstile
		if cfg.CloudflareSecretKey == "" {
			return nil, ErrProviderNotConfigured
		}

		// Site key is optional for verification but good to have for validation
		turnstile := NewTurnstileVerifier(cfg.CloudflareSecretKey, cfg.CloudflareSiteKey)
		return turnstile, nil

	default:
		return nil, ErrUnsupportedProvider
	}
}