package account

import (
	"go.uber.org/fx"
)

// AccountDomainModule contains all account domain repositories for dependency injection
var AccountDomainModule = fx.Options(
	fx.Provide(
		NewAccountRepo,
		NewEmailVerificationTokenRepo,
		NewPhoneNumberVerificationTokenRepo,
	),
)