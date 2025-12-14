package account

import (
	"log"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// AccountDomainModule contains all account domain repositories and services for dependency injection
var AccountDomainModule = fx.Options(
	fx.Provide(
		NewAccountRepo,
		NewEmailVerificationTokenRepo,
		NewPhoneNumberVerificationTokenRepo,
		NewAccountService,
		NewDummyMessageSenderForFX,
	),
)

// NewDummyMessageSenderForFX creates a new dummy message sender for FX dependency injection
func NewDummyMessageSenderForFX(logger *zap.Logger) MessageSender {
	// Convert zap logger to standard logger for compatibility with existing code
	stdLogger := log.Default()

	config := &DummySMSConfig{
		Enabled:      true,
		LogMessages:  true,
		FromNumber:   "+12345678901",
		ValidateOnly: false,
	}

	return NewDummyMessageSender(config, stdLogger)
}
