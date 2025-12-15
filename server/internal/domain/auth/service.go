package auth

import "server/internal/domain/account"

type AuthService struct {
	accountRepo                          *account.AccountRepo
	sessionRepo                          *SessionRepo
	emailVerificationTokenRepo           *account.EmailVerificationTokenRepo
	passwordResetTokenRepo               *PasswordResetTokenRepo
	webAuthnCredentialRepo               *WebAuthnCredentialRepo
	oauthCredentialRepo                  *OAuthCredentialRepo
	twoFactorAuthenticationChallengeRepo *TwoFactorAuthenticationChallengeRepo
	recoveryCodeRepo                     *RecoveryCodeRepo
	tempTwoFactorChallengeRepo           *TemporaryTwoFactorChallengeRepo
}

func NewAuthService(
	accountRepo *account.AccountRepo,
	sessionRepo *SessionRepo,
	emailVerificationTokenRepo *account.EmailVerificationTokenRepo,
	passwordResetTokenRepo *PasswordResetTokenRepo,
	webAuthnCredentialRepo *WebAuthnCredentialRepo,
	oauthCredentialRepo *OAuthCredentialRepo,
	twoFactorAuthenticationChallengeRepo *TwoFactorAuthenticationChallengeRepo,
	recoveryCodeRepo *RecoveryCodeRepo,
	tempTwoFactorChallengeRepo *TemporaryTwoFactorChallengeRepo,
) *AuthService {
	return &AuthService{
		accountRepo:                          accountRepo,
		sessionRepo:                          sessionRepo,
		emailVerificationTokenRepo:           emailVerificationTokenRepo,
		passwordResetTokenRepo:               passwordResetTokenRepo,
		webAuthnCredentialRepo:               webAuthnCredentialRepo,
		oauthCredentialRepo:                  oauthCredentialRepo,
		twoFactorAuthenticationChallengeRepo: twoFactorAuthenticationChallengeRepo,
		recoveryCodeRepo:                     recoveryCodeRepo,
		tempTwoFactorChallengeRepo:           tempTwoFactorChallengeRepo,
	}
}
