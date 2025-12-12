package account

import (
	"server/internal/domain/core"
	"time"

	"github.com/uptrace/bun"
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

	FullName        string   `bun:"full_name,notnull"`
	Email           string   `bun:"email,unique,notnull"`
	PasswordHash    string   `bun:"password_hash,nullzero"`
	TwoFactorSecret string   `bun:"two_factor_secret,nullzero"`
	AvatarURL       string   `bun:"avatar_url,nullzero"`
	AuthProviders   []string `bun:"auth_providers,array"`

	TermsAndPolicy TermsAndPolicy      `bun:"embed:terms_and_policy_"`
	AnalyticsPref  AnalyticsPreference `bun:"embed:analytics_pref_"`
}
