package captcha

import (
	"go.uber.org/fx"

	appconfig "server/internal/config"
)

// ProviderModule provides the captcha client and related services to the dependency injection container
var ProviderModule = fx.Module("captcha",
	// Provide captcha client as a constructor
	fx.Provide(NewCaptchaClient),

	// Provide captcha verifier interface (for dependency injection)
	fx.Provide(NewCaptchaClientProvider),
)

// CaptchaClientOptions provides configuration options for the captcha client
type CaptchaClientOptions struct {
	fx.In

	Config *appconfig.Config
}

// CaptchaVerifierOptions provides options for creating a captcha verifier
type CaptchaVerifierOptions struct {
	fx.In

	Config *appconfig.Config
}

// NewCaptchaClientFx is an fx-compatible constructor for CaptchaClient
func NewCaptchaClientFx(opts CaptchaClientOptions) (*CaptchaClient, error) {
	return NewCaptchaClient(opts.Config)
}

// NewCaptchaVerifierFx is an fx-compatible constructor that returns the interface
func NewCaptchaVerifierFx(opts CaptchaVerifierOptions) (CaptchaVerifier, error) {
	return NewCaptchaClientProvider(opts.Config)
}

// CaptchaClientGroup groups captcha-related dependencies together
type CaptchaClientGroup struct {
	fx.In

	Client   *CaptchaClient
	Verifier CaptchaVerifier `optional:"true"`
}

// ProvideCaptchaGroup provides a grouped set of captcha services
func ProvideCaptchaGroup(opts CaptchaClientOptions) (CaptchaClientGroup, error) {
	client, err := NewCaptchaClient(opts.Config)
	if err != nil {
		return CaptchaClientGroup{}, err
	}

	verifier, err := NewCaptchaClientProvider(opts.Config)
	if err != nil {
		// If verifier creation fails, we can still provide the client
		return CaptchaClientGroup{
			Client:   client,
			Verifier: nil,
		}, nil
	}

	return CaptchaClientGroup{
		Client:   client,
		Verifier: verifier,
	}, nil
}

// CaptchaProvider provides different ways to inject captcha services
var CaptchaProvider = fx.Options(
	fx.Provide(
		// Direct client provider
		NewCaptchaClientFx,

		// Interface provider for loose coupling
		NewCaptchaVerifierFx,

		// Group provider for multiple captcha services
		ProvideCaptchaGroup,
	),
)
