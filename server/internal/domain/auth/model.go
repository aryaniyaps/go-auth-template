package auth

import (
	"server/internal/domain/account"
	"server/internal/domain/core"

	"github.com/uptrace/bun"
)

type Session struct {
	core.CoreModel
	bun.BaseModel `bun:"table:sessions,alias:ses"`

	TokenHash string `bun:"token_hash,unique,notnull"`
	UserAgent string `bun:"user_agent,notnull"`
	IPAddress string `bun:"ip_address,notnull"`
	ExpiresAt int64  `bun:"expires_at,notnull"`
	AccountId int64  `bun:"account_id,notnull"`

	// account relationship
	Account *account.Account `bun:"rel:belongs-to,join:account_id=id"`
}

type PasswordResetToken struct {
	core.CoreModel
	bun.BaseModel `bun:"table:password_reset_tokens,alias:prt"`

	TokenHash string `bun:"token_hash,unique,notnull"`
	ExpiresAt int64  `bun:"expires_at,notnull"`
	AccountId int64  `bun:"account_id,notnull"`

	// account relationship
	Account *account.Account `bun:"rel:belongs-to,join:account_id=id"`
}

type WebAuthnCredential struct {
	core.CoreModel
	bun.BaseModel `bun:"table:webauthn_credentials,alias:wac"`

	CredentialID []byte `bun:"credential_id,unique,notnull"`
	PublicKey    []byte `bun:"public_key,notnull"`
	SignCount    uint32 `bun:"sign_count,notnull"`
	DeviceType   string `bun:"device_type"`
	BackedUp     bool   `bun:"backed_up,notnull"`
	Nickname     string `bun:"nickname"`
	Transports   string `bun:"transports,array,notnull"`
	AccountId    int64  `bun:"account_id,notnull"`

	// account relationship
	Account *account.Account `bun:"rel:belongs-to,join:account_id=id"`
}

type WebAuthnChallenge struct {
	core.CoreModel
	bun.BaseModel `bun:"table:webauthn_challenges,alias:wach"`

	Challenge          []byte `bun:"challenge,notnull,unique"`
	ExpiresAt          int64  `bun:"expires_at,notnull"`
	GeneratedAccountId int64  `bun:"generated_account_id,notnull"`
}

type OAuthCredential struct {
	core.CoreModel
	bun.BaseModel `bun:"table:oauth_credentials,alias:oac"`

	Provider       string `bun:"provider,notnull"`
	ProviderUserID string `bun:"provider_user_id,notnull"`
	// AccessToken    string `bun:"access_token,notnull"`
	// RefreshToken   string `bun:"refresh_token"`
	AccountId int64 `bun:"account_id,notnull"`

	// account relationship
	Account *account.Account `bun:"rel:belongs-to,join:account_id=id"`
}

type TwoFactorAuthenticationChallenge struct {
	core.CoreModel
	bun.BaseModel `bun:"table:two_factor_authentication_challenges,alias:tfac"`

	ChallengeHash string `bun:"challenge_hash,notnull,unique"`
	ExpiresAt     int64  `bun:"expires_at,notnull"`
	TOTPSecret    string `bun:"totp_secret,notnull"`
	AccountId     int64  `bun:"account_id,notnull"`

	// account relationship
	Account *account.Account `bun:"rel:belongs-to,join:account_id=id"`
}

type RecoveryCode struct {
	core.CoreModel
	bun.BaseModel `bun:"table:recovery_codes,alias:rc"`

	CodeHash  string `bun:"code_hash,notnull,unique"`
	AccountId int64  `bun:"account_id,notnull"`

	// account relationship
	Account *account.Account `bun:"rel:belongs-to,join:account_id=id"`
}

// temporary 2fa challenge used for password resets

type TemporaryTwoFactorChallenge struct {
	core.CoreModel
	bun.BaseModel `bun:"table:temporary_two_factor_challenges,alias:ttfc"`

	ChallengeHash      string `bun:"challenge_hash,notnull,unique"`
	ExpiresAt          int64  `bun:"expires_at,notnull"`
	PasswordResetToken string `bun:"password_reset_token,notnull"`
	AccountId          int64  `bun:"account_id,notnull"`

	// account relationship
	Account *account.Account `bun:"rel:belongs-to,join:account_id=id"`
}
