package auth

import (
	"go.uber.org/fx"
)

// AuthDomainModule contains all auth domain repositories for dependency injection
var AuthDomainModule = fx.Options(
	fx.Provide(
		NewSessionRepo,
		NewPasswordResetTokenRepo,
		NewWebAuthnCredentialRepo,
		NewWebAuthnChallengeRepo,
		NewOAuthCredentialRepo,
		NewTwoFactorAuthenticationChallengeRepo,
		NewRecoveryCodeRepo,
		NewTemporaryTwoFactorChallengeRepo,
	),
)