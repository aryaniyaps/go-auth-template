package account

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"server/internal/domain/core"
	"time"

	"github.com/uptrace/bun"
)

type TwoFactorProvider string

const (
	TwoFactorProviderAuthenticator TwoFactorProvider = "authenticator"
)

type TermsAndPolicy struct {
	Type      string    `bun:"type,notnull"` // e.g., "accepted", "updated"
	UpdatedAt time.Time `bun:"updated_at,nullzero"`
	Version   string    `bun:"version"`
}

type AnalyticsPreference struct {
	Type      string    `bun:"type,notnull"` // e.g., "enabled", "disabled"
	UpdatedAt time.Time `bun:"updated_at,nullzero"`
}

type Account struct {
	core.CoreModel
	bun.BaseModel `bun:"table:accounts,alias:acc"`

	FullName             string   `bun:"full_name,notnull"`
	Email                string   `bun:"email,unique,notnull"`
	PasswordHash         *string  `bun:"password_hash"`         // nullable for OAuth accounts
	TwoFactorSecret      *string  `bun:"two_factor_secret"`     // nullable
	InternalAvatarURL    *string  `bun:"avatar_url"`            // nullable
	AuthProviders        []string `bun:"auth_providers,array"`
	PhoneNumber          *string  `bun:"phone_number,unique"`   // nullable, unique constraint
	Profile              *string  `bun:"profile"`               // nullable, for future profile data
	WhatsAppJobAlerts    *bool    `bun:"whatsapp_job_alerts"`    // nullable, default false

	TermsAndPolicy TermsAndPolicy      `bun:"embed:terms_and_policy_"`
	AnalyticsPref  AnalyticsPreference `bun:"embed:analytics_pref_"`
}

func (a *Account) AvatarURL() string {
	if a.InternalAvatarURL != nil {
		return *a.InternalAvatarURL
	}
	h := md5.Sum([]byte(a.FullName))
	seedHash := hex.EncodeToString(h[:])

	// return f"https://api.dicebear.com/9.x/shapes/png?seed={seed_hash}"
	return fmt.Sprintf("https://api.dicebear.com/9.x/shapes/png?seed=%s", seedHash)
}

func (a *Account) Has2FAEnabled() bool {
	return a.TwoFactorSecret != nil && *a.TwoFactorSecret != ""
}

func (a *Account) TwoFactorProviders() []TwoFactorProvider {
	providers := make([]TwoFactorProvider, 0)
	if a.TwoFactorSecret != nil && *a.TwoFactorSecret != "" {
		providers = append(providers, TwoFactorProviderAuthenticator)
	}
	return providers
}

type EmailVerificationToken struct {
	core.CoreModel
	bun.BaseModel `bun:"table:email_verification_tokens,alias:evt"`

	Email     string    `bun:"email,notnull,unique"`
	TokenHash string    `bun:"token_hash,notnull"`
	ExpiresAt time.Time `bun:"expires_at,notnull"`
}

func (evt *EmailVerificationToken) IsExpired() bool {
	return time.Now().After(evt.ExpiresAt)
}

type PhoneNumberVerificationToken struct {
	core.CoreModel
	bun.BaseModel `bun:"table:phone_verification_tokens,alias:pvt"`

	PhoneNumber string    `bun:"phone_number,notnull,unique"`
	TokenHash   string    `bun:"token_hash,notnull"`
	ExpiresAt   time.Time `bun:"expires_at,notnull"`
}

func (pvt *PhoneNumberVerificationToken) IsExpired() bool {
	return time.Now().After(pvt.ExpiresAt)
}
