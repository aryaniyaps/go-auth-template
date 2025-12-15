package email

import (
	"go.uber.org/fx"

	appconfig "server/internal/config"
)

// ProviderModule provides the email client and related services to the dependency injection container
var ProviderModule = fx.Module("email",
	// Provide email client as a constructor
	fx.Provide(NewEmailClient),

	// Provide email sender interface (for dependency injection)
	fx.Provide(NewEmailClientProvider),
)

// EmailClientOptions provides configuration options for the email client
type EmailClientOptions struct {
	fx.In

	Config *appconfig.Config
}

// NewEmailClientFx is an fx-compatible constructor for EmailClient
func NewEmailClientFx(opts EmailClientOptions) (*EmailClient, error) {
	return NewEmailClient(opts.Config)
}

// EmailSenderOptions provides options for creating an email sender
type EmailSenderOptions struct {
	fx.In

	Config *appconfig.Config
}

// NewEmailSenderFx is an fx-compatible constructor that returns the interface
func NewEmailSenderFx(opts EmailSenderOptions) (EmailSender, error) {
	return NewEmailClientProvider(opts.Config)
}

// EmailClientGroup groups email-related dependencies together
type EmailClientGroup struct {
	fx.In

	Client *EmailClient
	Sender EmailSender `optional:"true"`
}

// ProvideEmailGroup provides a grouped set of email services
func ProvideEmailGroup(opts EmailClientOptions) (EmailClientGroup, error) {
	client, err := NewEmailClient(opts.Config)
	if err != nil {
		return EmailClientGroup{}, err
	}

	sender, err := NewEmailClientProvider(opts.Config)
	if err != nil {
		// If sender creation fails, we can still provide the client
		return EmailClientGroup{
			Client: client,
			Sender: nil,
		}, nil
	}

	return EmailClientGroup{
		Client: client,
		Sender: sender,
	}, nil
}

// EmailProvider provides different ways to inject email services
var EmailProvider = fx.Options(
	fx.Provide(
		// Direct client provider
		NewEmailClientFx,

		// Interface provider for loose coupling
		NewEmailSenderFx,

		// Group provider for multiple email services
		ProvideEmailGroup,
	),
)