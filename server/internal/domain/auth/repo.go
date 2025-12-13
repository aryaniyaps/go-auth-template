package auth

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"server/internal/infrastructure/db"

	"github.com/pquerna/otp/totp"
	"github.com/uptrace/bun"
)

// Import pagination infrastructure

// Error types for auth repository operations
var (
	ErrSessionNotFound                 = errors.New("session not found")
	ErrTokenExpired                    = errors.New("token has expired")
	ErrInvalidCredentials              = errors.New("invalid credentials")
	ErrWebAuthnCredentialNotFound      = errors.New("webauthn credential not found")
	ErrChallengeNotFound               = errors.New("challenge not found")
	ErrRecoveryCodeInvalid             = errors.New("recovery code invalid")
	ErrOAuthCredentialAlreadyExists    = errors.New("oauth credential already exists")
	ErrPasswordResetTokenNotFound      = errors.New("password reset token not found")
	ErrTwoFactorAuthenticationNotFound = errors.New("two factor authentication challenge not found")
	ErrTemporaryTwoFactorNotFound      = errors.New("temporary two factor challenge not found")
)

// Security utility functions for token generation and hashing

// generateSecureToken generates a cryptographically secure random hex token
func generateSecureToken(length int) (string, error) {
	if length <= 0 {
		length = 32 // default length
	}

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate secure token: %w", err)
	}

	return hex.EncodeToString(bytes), nil
}

// hashTokenMD5 hashes a token using MD5 (matching Python hashlib.md5)
func hashTokenMD5(token string) string {
	hash := md5.Sum([]byte(token))
	return hex.EncodeToString(hash[:])
}

// generateRecoveryCode generates an 8-character recovery code
func generateRecoveryCode() (string, error) {
	const charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate recovery code: %w", err)
	}

	for i, b := range bytes {
		bytes[i] = charset[b%byte(len(charset))]
	}

	return string(bytes), nil
}

// generateTwoFactorSecret generates a TOTP secret
func generateTwoFactorSecret() (string, error) {
	secret, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "auth-system",
		AccountName: "user@example.com", // Will be updated per user
		SecretSize:  32,
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate 2FA secret: %w", err)
	}
	return secret.Secret(), nil
}

// SessionRepo interface defines methods for session management
type SessionRepo interface {
	Create(ctx context.Context, accountId int64, userAgent string, ipAddress string) (string, error)
	Get(ctx context.Context, token string, fetchAccount bool) (*Session, error)
	GetBySessionAccountId(ctx context.Context, sessionId int64, accountId int64, exceptSessionToken string) (*Session, error)
	GetAllList(ctx context.Context, accountId int64, exceptSessionToken string) ([]*Session, error)
	GetAllByAccountId(ctx context.Context, accountId int64, exceptSessionToken string, first *int, last *int, before *string, after *string) (*db.PaginatedResult[*Session, int64], error)
	DeleteByToken(ctx context.Context, token string) error
	Delete(ctx context.Context, session *Session) error
	DeleteMany(ctx context.Context, sessionIds []int64) error
	DeleteAll(ctx context.Context, accountId int64) error

	// Static methods for token operations
	GenerateSessionToken() (string, error)
	HashSessionToken(token string) string
}

// Session repository implementation
type sessionRepo struct {
	db *bun.DB
}

func NewSessionRepo(db *bun.DB) SessionRepo {
	return &sessionRepo{db: db}
}

// Static methods
func (r *sessionRepo) GenerateSessionToken() (string, error) {
	return generateSecureToken(32)
}

func (r *sessionRepo) HashSessionToken(token string) string {
	return hashTokenMD5(token)
}

// Session management
func (r *sessionRepo) Create(ctx context.Context, accountId int64, userAgent string, ipAddress string) (string, error) {
	sessionToken, err := r.GenerateSessionToken()
	if err != nil {
		return "", fmt.Errorf("failed to generate session token: %w", err)
	}

	expiresAt := time.Now().Add(24 * time.Hour) // 24 hour session
	session := &Session{
		TokenHash: r.HashSessionToken(sessionToken),
		UserAgent: userAgent,
		IPAddress: ipAddress,
		ExpiresAt: expiresAt.Unix(),
		AccountId: accountId,
	}

	_, err = r.db.NewInsert().
		Model(session).
		Returning("*").
		Exec(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	return sessionToken, nil
}

func (r *sessionRepo) Get(ctx context.Context, token string, fetchAccount bool) (*Session, error) {
	session := &Session{}
	query := r.db.NewSelect().
		Model(session).
		Where("token_hash = ?", r.HashSessionToken(token))

	if fetchAccount {
		query = query.Relation("Account")
	}

	err := query.Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Check if session is expired
	if time.Now().Unix() > session.ExpiresAt {
		return nil, ErrTokenExpired
	}

	return session, nil
}

func (r *sessionRepo) GetBySessionAccountId(ctx context.Context, sessionId int64, accountId int64, exceptSessionToken string) (*Session, error) {
	session := &Session{}
	query := r.db.NewSelect().
		Model(session).
		Where("id = ?", sessionId).
		Where("account_id = ?", accountId)

	if exceptSessionToken != "" {
		query = query.Where("token_hash != ?", r.HashSessionToken(exceptSessionToken))
	}

	err := query.Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to get session by ID and account ID: %w", err)
	}

	return session, nil
}

func (r *sessionRepo) GetAllList(ctx context.Context, accountId int64, exceptSessionToken string) ([]*Session, error) {
	sessions := make([]*Session, 0)
	query := r.db.NewSelect().
		Model(&sessions).
		Where("account_id = ?", accountId)

	if exceptSessionToken != "" {
		query = query.Where("token_hash != ?", r.HashSessionToken(exceptSessionToken))
	}

	err := query.Order("created_at DESC").Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all sessions for account: %w", err)
	}

	return sessions, nil
}

func (r *sessionRepo) GetAllByAccountId(ctx context.Context, accountId int64, exceptSessionToken string, first *int, last *int, before *string, after *string) (*db.PaginatedResult[*Session, int64], error) {
	sessions := make([]*Session, 0)
	query := r.db.NewSelect().
		Model(&sessions).
		Where("account_id = ?", accountId)

	if exceptSessionToken != "" {
		query = query.Where("token_hash != ?", r.HashSessionToken(exceptSessionToken))
	}

	// Apply pagination using simplified API
	paginationOptions := db.PaginationOptions{
		First:  first,
		Last:   last,
		After:  after,
		Before: before,
	}

	// Validate pagination parameters
	if err := db.ValidatePagination(paginationOptions); err != nil {
		return nil, fmt.Errorf("invalid pagination parameters: %w", err)
	}

	query = db.ApplyPagination(query, paginationOptions)

	err := query.Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get paginated sessions: %w", err)
	}

	// Process pagination metadata - returns PaginatedResult with int64 cursors
	result := db.ProcessPaginatedResult[*Session, int64](sessions, first, last)
	return &result, nil
}

func (r *sessionRepo) DeleteByToken(ctx context.Context, token string) error {
	_, err := r.db.NewDelete().
		Model((*Session)(nil)).
		Where("token_hash = ?", r.HashSessionToken(token)).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete session by token: %w", err)
	}
	return nil
}

func (r *sessionRepo) Delete(ctx context.Context, session *Session) error {
	_, err := r.db.NewDelete().
		Model(session).
		Where("id = ?", session.ID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

func (r *sessionRepo) DeleteMany(ctx context.Context, sessionIds []int64) error {
	if len(sessionIds) == 0 {
		return nil
	}

	_, err := r.db.NewDelete().
		Model((*Session)(nil)).
		Where("id IN (?)", bun.In(sessionIds)).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete many sessions: %w", err)
	}
	return nil
}

func (r *sessionRepo) DeleteAll(ctx context.Context, accountId int64) error {
	_, err := r.db.NewDelete().
		Model((*Session)(nil)).
		Where("account_id = ?", accountId).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete all sessions for account: %w", err)
	}
	return nil
}

// PasswordResetTokenRepo interface defines methods for password reset token management
type PasswordResetTokenRepo interface {
	Create(ctx context.Context, accountId int64) (string, error)
	Get(ctx context.Context, token string, email string) (*PasswordResetToken, error)
	GetByAccount(ctx context.Context, accountId int64) (*PasswordResetToken, error)
	Delete(ctx context.Context, token *PasswordResetToken) error

	// Static methods for token operations
	GeneratePasswordResetToken() (string, error)
	HashPasswordResetToken(token string) string
}

// Password reset token repository implementation
type passwordResetTokenRepo struct {
	db *bun.DB
}

func NewPasswordResetTokenRepo(db *bun.DB) PasswordResetTokenRepo {
	return &passwordResetTokenRepo{db: db}
}

// Static methods
func (r *passwordResetTokenRepo) GeneratePasswordResetToken() (string, error) {
	return generateSecureToken(32)
}

func (r *passwordResetTokenRepo) HashPasswordResetToken(token string) string {
	return hashTokenMD5(token)
}

func (r *passwordResetTokenRepo) Create(ctx context.Context, accountId int64) (string, error) {
	token, err := r.GeneratePasswordResetToken()
	if err != nil {
		return "", fmt.Errorf("failed to generate password reset token: %w", err)
	}

	expiresAt := time.Now().Add(1 * time.Hour) // 1 hour expiry
	passwordResetToken := &PasswordResetToken{
		TokenHash: r.HashPasswordResetToken(token),
		ExpiresAt: expiresAt.Unix(),
		AccountId: accountId,
	}

	_, err = r.db.NewInsert().
		Model(passwordResetToken).
		Returning("*").
		Exec(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create password reset token: %w", err)
	}

	return token, nil
}

func (r *passwordResetTokenRepo) Get(ctx context.Context, token string, email string) (*PasswordResetToken, error) {
	passwordResetToken := &PasswordResetToken{}
	err := r.db.NewSelect().
		Model(passwordResetToken).
		Where("token_hash = ?", r.HashPasswordResetToken(token)).
		Relation("Account", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where("email = ?", email)
		}).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPasswordResetTokenNotFound
		}
		return nil, fmt.Errorf("failed to get password reset token: %w", err)
	}
	// Check if token is expired
	if time.Now().Unix() > passwordResetToken.ExpiresAt {
		return nil, ErrTokenExpired
	}

	return passwordResetToken, nil
}

func (r *passwordResetTokenRepo) GetByAccount(ctx context.Context, accountId int64) (*PasswordResetToken, error) {
	passwordResetToken := &PasswordResetToken{}
	err := r.db.NewSelect().
		Model(passwordResetToken).
		Where("account_id = ?", accountId).
		Relation("Account").
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPasswordResetTokenNotFound
		}
		return nil, fmt.Errorf("failed to get password reset token by account: %w", err)
	}

	return passwordResetToken, nil
}

func (r *passwordResetTokenRepo) Delete(ctx context.Context, token *PasswordResetToken) error {
	_, err := r.db.NewDelete().
		Model(token).
		Where("id = ?", token.ID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete password reset token: %w", err)
	}
	return nil
}

// WebAuthnCredentialRepo interface defines methods for WebAuthn credential management
type WebAuthnCredentialRepo interface {
	Create(ctx context.Context, accountId int64, credentialId []byte, credentialPublicKey []byte, signCount uint32, deviceType string, backedUp bool, transports []string, nickname string) (*WebAuthnCredential, error)
	UpdateSignCount(ctx context.Context, credentialId []byte, signCount uint32) error
	Get(ctx context.Context, credentialId []byte, fetchAccount bool) (*WebAuthnCredential, error)
	GetByAccountCredentialId(ctx context.Context, accountId int64, webAuthnCredentialId int64) (*WebAuthnCredential, error)
	Delete(ctx context.Context, credential *WebAuthnCredential) error
	GetAllByAccountList(ctx context.Context, accountId int64) ([]*WebAuthnCredential, error)
	Update(ctx context.Context, webAuthnCredentialId int64, nickname string) (*WebAuthnCredential, error)
	GetAllByAccountId(ctx context.Context, accountId int64, first *int, last *int, before *string, after *string) (*db.PaginatedResult[*WebAuthnCredential, int64], error)
}

// WebAuthn credential repository implementation
type webAuthnCredentialRepo struct {
	db *bun.DB
}

func NewWebAuthnCredentialRepo(db *bun.DB) WebAuthnCredentialRepo {
	return &webAuthnCredentialRepo{db: db}
}

func (r *webAuthnCredentialRepo) Create(ctx context.Context, accountId int64, credentialId []byte, credentialPublicKey []byte, signCount uint32, deviceType string, backedUp bool, transports []string, nickname string) (*WebAuthnCredential, error) {
	webAuthnCredential := &WebAuthnCredential{
		CredentialID: credentialId,
		PublicKey:    credentialPublicKey,
		SignCount:    signCount,
		DeviceType:   deviceType,
		BackedUp:     backedUp,
		Transports:   transports,
		Nickname:     nickname,
		AccountId:    accountId,
	}

	_, err := r.db.NewInsert().
		Model(webAuthnCredential).
		Returning("*").
		Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create webauthn credential: %w", err)
	}

	return webAuthnCredential, nil
}

func (r *webAuthnCredentialRepo) UpdateSignCount(ctx context.Context, credentialId []byte, signCount uint32) error {
	_, err := r.db.NewUpdate().
		Model((*WebAuthnCredential)(nil)).
		Set("sign_count = ?", signCount).
		Set("updated_at = ?", time.Now()).
		Where("credential_id = ?", credentialId).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update sign count: %w", err)
	}
	return nil
}

func (r *webAuthnCredentialRepo) Get(ctx context.Context, credentialId []byte, fetchAccount bool) (*WebAuthnCredential, error) {
	webAuthnCredential := &WebAuthnCredential{}
	query := r.db.NewSelect().
		Model(webAuthnCredential).
		Where("credential_id = ?", credentialId)

	if fetchAccount {
		query = query.Relation("Account")
	}

	err := query.Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrWebAuthnCredentialNotFound
		}
		return nil, fmt.Errorf("failed to get webauthn credential: %w", err)
	}

	return webAuthnCredential, nil
}

func (r *webAuthnCredentialRepo) GetByAccountCredentialId(ctx context.Context, accountId int64, webAuthnCredentialId int64) (*WebAuthnCredential, error) {
	webAuthnCredential := &WebAuthnCredential{}
	err := r.db.NewSelect().
		Model(webAuthnCredential).
		Where("id = ?", webAuthnCredentialId).
		Where("account_id = ?", accountId).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrWebAuthnCredentialNotFound
		}
		return nil, fmt.Errorf("failed to get webauthn credential by ID: %w", err)
	}

	return webAuthnCredential, nil
}

func (r *webAuthnCredentialRepo) Delete(ctx context.Context, credential *WebAuthnCredential) error {
	_, err := r.db.NewDelete().
		Model(credential).
		Where("id = ?", credential.ID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete webauthn credential: %w", err)
	}
	return nil
}

func (r *webAuthnCredentialRepo) GetAllByAccountList(ctx context.Context, accountId int64) ([]*WebAuthnCredential, error) {
	credentials := make([]*WebAuthnCredential, 0)
	err := r.db.NewSelect().
		Model(&credentials).
		Where("account_id = ?", accountId).
		Order("created_at DESC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all webauthn credentials for account: %w", err)
	}

	return credentials, nil
}

func (r *webAuthnCredentialRepo) Update(ctx context.Context, webAuthnCredentialId int64, nickname string) (*WebAuthnCredential, error) {
	// First get the existing credential
	credential := &WebAuthnCredential{}
	err := r.db.NewSelect().
		Model(credential).
		Where("id = ?", webAuthnCredentialId).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get webauthn credential for update: %w", err)
	}

	credential.Nickname = nickname

	_, err = r.db.NewUpdate().
		Model(credential).
		Set("nickname = ?", nickname).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", webAuthnCredentialId).
		Returning("*").
		Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update webauthn credential: %w", err)
	}

	return credential, nil
}

func (r *webAuthnCredentialRepo) GetAllByAccountId(ctx context.Context, accountId int64, first *int, last *int, before *string, after *string) (*db.PaginatedResult[*WebAuthnCredential, int64], error) {
	credentials := make([]*WebAuthnCredential, 0)
	query := r.db.NewSelect().
		Model(&credentials).
		Where("account_id = ?", accountId)

	// Apply pagination using simplified API
	paginationOptions := db.PaginationOptions{
		First:  first,
		Last:   last,
		After:  after,
		Before: before,
	}

	// Validate pagination parameters
	if err := db.ValidatePagination(paginationOptions); err != nil {
		return nil, fmt.Errorf("invalid pagination parameters: %w", err)
	}

	query = db.ApplyPagination(query, paginationOptions)

	err := query.Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get paginated webauthn credentials: %w", err)
	}

	// Process pagination metadata - returns PaginatedResult with string cursors
	result := db.ProcessPaginatedResult[*WebAuthnCredential, int64](credentials, first, last)
	return &result, nil
}

// WebAuthnChallengeRepo interface defines methods for WebAuthn challenge management
type WebAuthnChallengeRepo interface {
	Create(ctx context.Context, challenge []byte, generatedAccountId int64) (*WebAuthnChallenge, error)
	Get(ctx context.Context, challenge []byte) (*WebAuthnChallenge, error)
	Delete(ctx context.Context, webauthnChallenge *WebAuthnChallenge) error
}

// WebAuthn challenge repository implementation
type webAuthnChallengeRepo struct {
	db *bun.DB
}

func NewWebAuthnChallengeRepo(db *bun.DB) WebAuthnChallengeRepo {
	return &webAuthnChallengeRepo{db: db}
}

func (r *webAuthnChallengeRepo) Create(ctx context.Context, challenge []byte, generatedAccountId int64) (*WebAuthnChallenge, error) {
	expiresAt := time.Now().Add(5 * time.Minute) // 5 minute expiry
	webauthnChallenge := &WebAuthnChallenge{
		Challenge:          challenge,
		ExpiresAt:          expiresAt.Unix(),
		GeneratedAccountId: generatedAccountId,
	}

	_, err := r.db.NewInsert().
		Model(webauthnChallenge).
		Returning("*").
		Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create webauthn challenge: %w", err)
	}

	return webauthnChallenge, nil
}

func (r *webAuthnChallengeRepo) Get(ctx context.Context, challenge []byte) (*WebAuthnChallenge, error) {
	webauthnChallenge := &WebAuthnChallenge{}
	err := r.db.NewSelect().
		Model(webauthnChallenge).
		Where("challenge = ?", challenge).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrChallengeNotFound
		}
		return nil, fmt.Errorf("failed to get webauthn challenge: %w", err)
	}
	// Check if challenge is expired
	if time.Now().Unix() > webauthnChallenge.ExpiresAt {
		return nil, ErrTokenExpired
	}

	return webauthnChallenge, nil
}

func (r *webAuthnChallengeRepo) Delete(ctx context.Context, webauthnChallenge *WebAuthnChallenge) error {
	_, err := r.db.NewDelete().
		Model(webauthnChallenge).
		Where("id = ?", webauthnChallenge.ID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete webauthn challenge: %w", err)
	}
	return nil
}

// OAuthCredentialRepo interface defines methods for OAuth credential management
type OAuthCredentialRepo interface {
	Create(ctx context.Context, accountId int64, provider string, providerUserId string) (*OAuthCredential, error)
	GetByProviderUser(ctx context.Context, provider string, providerUserId string, fetchAccount bool) (*OAuthCredential, error)
	GetByAccountProvider(ctx context.Context, accountId int64, provider string, fetchAccount bool) (*OAuthCredential, error)
	Delete(ctx context.Context, credential *OAuthCredential) error
}

// OAuth credential repository implementation
type oAuthCredentialRepo struct {
	db *bun.DB
}

func NewOAuthCredentialRepo(db *bun.DB) OAuthCredentialRepo {
	return &oAuthCredentialRepo{db: db}
}

func (r *oAuthCredentialRepo) Create(ctx context.Context, accountId int64, provider string, providerUserId string) (*OAuthCredential, error) {
	oauthCredential := &OAuthCredential{
		Provider:       provider,
		ProviderUserID: providerUserId,
		AccountId:      accountId,
	}

	_, err := r.db.NewInsert().
		Model(oauthCredential).
		Returning("*").
		Exec(ctx)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrOAuthCredentialAlreadyExists
		}
		return nil, fmt.Errorf("failed to create oauth credential: %w", err)
	}
	return oauthCredential, nil
}

func (r *oAuthCredentialRepo) GetByProviderUser(ctx context.Context, provider string, providerUserId string, fetchAccount bool) (*OAuthCredential, error) {
	oauthCredential := &OAuthCredential{}
	query := r.db.NewSelect().
		Model(oauthCredential).
		Where("provider = ?", provider).
		Where("provider_user_id = ?", providerUserId)

	if fetchAccount {
		query = query.Relation("Account")
	}

	err := query.Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrOAuthCredentialAlreadyExists
		}
		return nil, fmt.Errorf("failed to get oauth credential by provider user: %w", err)
	}

	return oauthCredential, nil
}

func (r *oAuthCredentialRepo) GetByAccountProvider(ctx context.Context, accountId int64, provider string, fetchAccount bool) (*OAuthCredential, error) {
	oauthCredential := &OAuthCredential{}
	query := r.db.NewSelect().
		Model(oauthCredential).
		Where("account_id = ?", accountId).
		Where("provider = ?", provider)

	if fetchAccount {
		query = query.Relation("Account")
	}

	err := query.Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrOAuthCredentialAlreadyExists
		}
		return nil, fmt.Errorf("failed to get oauth credential by account provider: %w", err)
	}

	return oauthCredential, nil
}

func (r *oAuthCredentialRepo) Delete(ctx context.Context, credential *OAuthCredential) error {
	_, err := r.db.NewDelete().
		Model(credential).
		Where("id = ?", credential.ID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete oauth credential: %w", err)
	}
	return nil
}

// TwoFactorAuthenticationChallengeRepo interface defines methods for 2FA challenge management
type TwoFactorAuthenticationChallengeRepo interface {
	Create(ctx context.Context, accountId int64, totpSecret string) (string, *TwoFactorAuthenticationChallenge, error)
	Get(ctx context.Context, challenge string, fetchAccount bool) (*TwoFactorAuthenticationChallenge, error)
	Delete(ctx context.Context, challenge *TwoFactorAuthenticationChallenge) error

	// Static methods for challenge operations
	GenerateChallenge() (string, error)
	HashChallenge(challenge string) string
	GenerateTwoFactorSecret() (string, error)
}

// Two-factor authentication challenge repository implementation
type twoFactorAuthenticationChallengeRepo struct {
	db *bun.DB
}

func NewTwoFactorAuthenticationChallengeRepo(db *bun.DB) TwoFactorAuthenticationChallengeRepo {
	return &twoFactorAuthenticationChallengeRepo{db: db}
}

// Static methods
func (r *twoFactorAuthenticationChallengeRepo) GenerateChallenge() (string, error) {
	return generateSecureToken(32)
}

func (r *twoFactorAuthenticationChallengeRepo) HashChallenge(challenge string) string {
	return hashTokenMD5(challenge)
}

func (r *twoFactorAuthenticationChallengeRepo) GenerateTwoFactorSecret() (string, error) {
	return generateTwoFactorSecret()
}

func (r *twoFactorAuthenticationChallengeRepo) Create(ctx context.Context, accountId int64, totpSecret string) (string, *TwoFactorAuthenticationChallenge, error) {
	challenge, err := r.GenerateChallenge()
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate challenge: %w", err)
	}

	var secret string
	if totpSecret != "" {
		secret = totpSecret
	} else {
		secret, err = r.GenerateTwoFactorSecret()
		if err != nil {
			return "", nil, fmt.Errorf("failed to generate 2FA secret: %w", err)
		}
	}

	expiresAt := time.Now().Add(5 * time.Minute) // 5 minute expiry
	twoFactorChallenge := &TwoFactorAuthenticationChallenge{
		ChallengeHash: r.HashChallenge(challenge),
		ExpiresAt:     expiresAt.Unix(),
		TOTPSecret:    secret,
		AccountId:     accountId,
	}

	_, err = r.db.NewInsert().
		Model(twoFactorChallenge).
		Returning("*").
		Exec(ctx)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create 2FA challenge: %w", err)
	}

	return challenge, twoFactorChallenge, nil
}

func (r *twoFactorAuthenticationChallengeRepo) Get(ctx context.Context, challenge string, fetchAccount bool) (*TwoFactorAuthenticationChallenge, error) {
	twoFactorChallenge := &TwoFactorAuthenticationChallenge{}
	query := r.db.NewSelect().
		Model(twoFactorChallenge).
		Where("challenge_hash = ?", r.HashChallenge(challenge))

	if fetchAccount {
		query = query.Relation("Account")
	}

	err := query.Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTwoFactorAuthenticationNotFound
		}
		return nil, fmt.Errorf("failed to get 2FA challenge: %w", err)
	}

	// Check if challenge is expired
	if time.Now().Unix() > twoFactorChallenge.ExpiresAt {
		return nil, ErrTokenExpired
	}

	return twoFactorChallenge, nil
}

func (r *twoFactorAuthenticationChallengeRepo) Delete(ctx context.Context, challenge *TwoFactorAuthenticationChallenge) error {
	_, err := r.db.NewDelete().
		Model(challenge).
		Where("id = ?", challenge.ID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete 2FA challenge: %w", err)
	}
	return nil
}

// RecoveryCodeRepo interface defines methods for recovery code management
type RecoveryCodeRepo interface {
	Create(ctx context.Context, accountId int64, code string) (string, error)
	CreateMany(ctx context.Context, accountId int64, codeCount int) ([]string, error)
	DeleteAll(ctx context.Context, accountId int64) error
	Delete(ctx context.Context, recoveryCode *RecoveryCode) error
	Get(ctx context.Context, accountId int64, code string) (*RecoveryCode, error)
	GetAllByAccountId(ctx context.Context, accountId int64) ([]*RecoveryCode, error)

	// Static methods for recovery code operations
	GenerateRecoveryCode() (string, error)
	HashRecoveryCode(code string) string
}

// Recovery code repository implementation
type recoveryCodeRepo struct {
	db *bun.DB
}

func NewRecoveryCodeRepo(db *bun.DB) RecoveryCodeRepo {
	return &recoveryCodeRepo{db: db}
}

// Static methods
func (r *recoveryCodeRepo) GenerateRecoveryCode() (string, error) {
	return generateRecoveryCode()
}

func (r *recoveryCodeRepo) HashRecoveryCode(code string) string {
	return hashTokenMD5(code)
}

func (r *recoveryCodeRepo) Create(ctx context.Context, accountId int64, code string) (string, error) {
	if code == "" {
		var err error
		code, err = r.GenerateRecoveryCode()
		if err != nil {
			return "", fmt.Errorf("failed to generate recovery code: %w", err)
		}
	}

	recoveryCode := &RecoveryCode{
		CodeHash:  r.HashRecoveryCode(code),
		AccountId: accountId,
	}

	_, err := r.db.NewInsert().
		Model(recoveryCode).
		Exec(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create recovery code: %w", err)
	}

	return code, nil
}

func (r *recoveryCodeRepo) CreateMany(ctx context.Context, accountId int64, codeCount int) ([]string, error) {
	if codeCount <= 0 {
		codeCount = 10
	}

	codes := make([]string, 0, codeCount)
	recoveryCodes := make([]*RecoveryCode, 0, codeCount)

	for i := 0; i < codeCount; i++ {
		code, err := r.GenerateRecoveryCode()
		if err != nil {
			return nil, fmt.Errorf("failed to generate recovery code: %w", err)
		}

		codes = append(codes, code)
		recoveryCodes = append(recoveryCodes, &RecoveryCode{
			CodeHash:  r.HashRecoveryCode(code),
			AccountId: accountId,
		})
	}

	_, err := r.db.NewInsert().
		Model(&recoveryCodes).
		Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create recovery codes: %w", err)
	}

	return codes, nil
}

func (r *recoveryCodeRepo) DeleteAll(ctx context.Context, accountId int64) error {
	_, err := r.db.NewDelete().
		Model((*RecoveryCode)(nil)).
		Where("account_id = ?", accountId).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete all recovery codes for account: %w", err)
	}
	return nil
}

func (r *recoveryCodeRepo) Delete(ctx context.Context, recoveryCode *RecoveryCode) error {
	_, err := r.db.NewDelete().
		Model(recoveryCode).
		Where("id = ?", recoveryCode.ID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete recovery code: %w", err)
	}
	return nil
}

func (r *recoveryCodeRepo) Get(ctx context.Context, accountId int64, code string) (*RecoveryCode, error) {
	recoveryCode := &RecoveryCode{}
	err := r.db.NewSelect().
		Model(recoveryCode).
		Where("account_id = ?", accountId).
		Where("code_hash = ?", r.HashRecoveryCode(code)).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecoveryCodeInvalid
		}
		return nil, fmt.Errorf("failed to get recovery code: %w", err)
	}

	return recoveryCode, nil
}

func (r *recoveryCodeRepo) GetAllByAccountId(ctx context.Context, accountId int64) ([]*RecoveryCode, error) {
	recoveryCodes := make([]*RecoveryCode, 0)
	err := r.db.NewSelect().
		Model(&recoveryCodes).
		Where("account_id = ?", accountId).
		Order("created_at DESC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all recovery codes for account: %w", err)
	}

	return recoveryCodes, nil
}

// TemporaryTwoFactorChallengeRepo interface defines methods for temporary 2FA challenge management
type TemporaryTwoFactorChallengeRepo interface {
	Create(ctx context.Context, accountId int64, passwordResetTokenId int64) (string, *TemporaryTwoFactorChallenge, error)
	Get(ctx context.Context, challenge string, passwordResetTokenId int64, fetchAccount bool) (*TemporaryTwoFactorChallenge, error)
	Delete(ctx context.Context, temporaryTwoFactorChallenge *TemporaryTwoFactorChallenge) error

	// Static methods for challenge operations
	GenerateChallenge() (string, error)
	HashChallenge(challenge string) string
}

// Temporary two-factor challenge repository implementation
type temporaryTwoFactorChallengeRepo struct {
	db *bun.DB
}

func NewTemporaryTwoFactorChallengeRepo(db *bun.DB) TemporaryTwoFactorChallengeRepo {
	return &temporaryTwoFactorChallengeRepo{db: db}
}

// Static methods
func (r *temporaryTwoFactorChallengeRepo) GenerateChallenge() (string, error) {
	return generateSecureToken(32)
}

func (r *temporaryTwoFactorChallengeRepo) HashChallenge(challenge string) string {
	return hashTokenMD5(challenge)
}

func (r *temporaryTwoFactorChallengeRepo) Create(ctx context.Context, accountId int64, passwordResetTokenId int64) (string, *TemporaryTwoFactorChallenge, error) {
	challenge, err := r.GenerateChallenge()
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate temporary 2FA challenge: %w", err)
	}

	expiresAt := time.Now().Add(5 * time.Minute) // 5 minute expiry
	temporaryTwoFactorChallenge := &TemporaryTwoFactorChallenge{
		ChallengeHash:      r.HashChallenge(challenge),
		ExpiresAt:          expiresAt.Unix(),
		PasswordResetToken: fmt.Sprintf("%d", passwordResetTokenId), // Convert to string
		AccountId:          accountId,
	}

	_, err = r.db.NewInsert().
		Model(temporaryTwoFactorChallenge).
		Returning("*").
		Exec(ctx)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temporary 2FA challenge: %w", err)
	}

	return challenge, temporaryTwoFactorChallenge, nil
}

func (r *temporaryTwoFactorChallengeRepo) Get(ctx context.Context, challenge string, passwordResetTokenId int64, fetchAccount bool) (*TemporaryTwoFactorChallenge, error) {
	temporaryTwoFactorChallenge := &TemporaryTwoFactorChallenge{}
	query := r.db.NewSelect().
		Model(temporaryTwoFactorChallenge).
		Where("challenge_hash = ?", r.HashChallenge(challenge)).
		Where("password_reset_token = ?", fmt.Sprintf("%d", passwordResetTokenId))

	if fetchAccount {
		query = query.Relation("Account")
	}

	err := query.Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTemporaryTwoFactorNotFound
		}
		return nil, fmt.Errorf("failed to get temporary 2FA challenge: %w", err)
	}

	// Check if challenge is expired
	if time.Now().Unix() > temporaryTwoFactorChallenge.ExpiresAt {
		return nil, ErrTokenExpired
	}

	return temporaryTwoFactorChallenge, nil
}

func (r *temporaryTwoFactorChallengeRepo) Delete(ctx context.Context, temporaryTwoFactorChallenge *TemporaryTwoFactorChallenge) error {
	_, err := r.db.NewDelete().
		Model(temporaryTwoFactorChallenge).
		Where("id = ?", temporaryTwoFactorChallenge.ID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete temporary 2FA challenge: %w", err)
	}
	return nil
}

// Helper functions for error handling (reused from account repo)
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(strings.ToLower(errStr), "duplicate key") ||
		strings.Contains(strings.ToLower(errStr), "unique constraint") ||
		strings.Contains(strings.ToLower(errStr), "unique violation")
}
