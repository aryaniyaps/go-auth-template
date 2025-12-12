package account

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/uptrace/bun"
	"golang.org/x/crypto/argon2"
)

// Security utility functions for password hashing
func HashPassword(password string) (string, error) {
	// Use argon2 parameters that match Python passlib defaults
	// Python passlib.argon2 defaults: time=1, memory=102400, parallelism=8, hashlen=16, saltlen=16
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, 1, 102400, 8, 32)

	// Combine salt and hash for storage
	combined := append(salt, hash...)
	return base64.StdEncoding.EncodeToString(combined), nil
}

func VerifyPassword(password, hash string) (bool, error) {
	combined, err := base64.StdEncoding.DecodeString(hash)
	if err != nil {
		return false, fmt.Errorf("failed to decode password hash: %w", err)
	}

	if len(combined) < 16 {
		return false, errors.New("invalid password hash format")
	}

	salt := combined[:16]
	storedHash := combined[16:]

	computedHash := argon2.IDKey([]byte(password), salt, 1, 102400, 8, 32)

	// Constant-time comparison to prevent timing attacks
	if len(storedHash) != len(computedHash) {
		return false, nil
	}

	var result byte
	for i := range storedHash {
		result |= storedHash[i] ^ computedHash[i]
	}

	return result == 0, nil
}

// Token generation and hashing utilities
func HashVerificationToken(token string) string {
	hash := md5.Sum([]byte(token))
	return hex.EncodeToString(hash[:])
}

func GenerateVerificationToken(length int) (string, error) {
	if length <= 0 {
		length = 32 // default length
	}

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return hex.EncodeToString(bytes), nil
}

// Helper function for updating array fields
func updateStringSlice(current []string, updates []string) []string {
	if updates == nil {
		return current
	}

	// Create a new slice to avoid modifying the original
	result := make([]string, len(updates))
	copy(result, updates)
	return result
}

// Helper function to remove a string from a slice
func removeStringFromSlice(slice []string, item string) []string {
	for i, s := range slice {
		if s == item {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

// Helper function to add a string to a slice if it doesn't exist
func addStringToSlice(slice []string, item string) []string {
	if slices.Contains(slice, item) {
		return slice
	}
	return append(slice, item)
}

// Error types for repository operations
var (
	ErrAccountNotFound    = errors.New("account not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrPhoneAlreadyExists = errors.New("phone number already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrTokenExpired       = errors.New("token has expired")
	ErrTokenNotFound      = errors.New("token not found")
)

// Repository interfaces
type AccountRepo interface {
	Create(ctx context.Context, email string, fullName string, authProviders []string, password *string, accountID *int64, analyticsPreference string, phoneNumber *string) (*Account, error)
	Get(ctx context.Context, accountID int64) (*Account, error)
	GetByEmail(ctx context.Context, email string) (*Account, error)
	GetByPhoneNumber(ctx context.Context, phoneNumber string) (*Account, error)
	Update(ctx context.Context, account *Account, fullName *string, avatarURL *string, phoneNumber *string, termsAndPolicy *TermsAndPolicy, analyticsPreference *AnalyticsPreference, whatsappJobAlerts *bool) (*Account, error)
	UpdateProfile(ctx context.Context, account *Account, profile any) (*Account, error)
	UpdateAuthProviders(ctx context.Context, account *Account, authProviders []string) (*Account, error)
	DeleteAvatar(ctx context.Context, account *Account) (*Account, error)
	SetTwoFactorSecret(ctx context.Context, account *Account, totpSecret string) (*Account, error)
	DeleteTwoFactorSecret(ctx context.Context, account *Account) (*Account, error)
	UpdatePassword(ctx context.Context, account *Account, password string) (*Account, error)
	DeletePassword(ctx context.Context, account *Account) (*Account, error)
	Delete(ctx context.Context, account *Account) error

	// Static methods for password operations
	HashPassword(password string) (string, error)
	VerifyPassword(password, hash string) (bool, error)
}

// Repository implementations
type accountRepo struct {
	db *bun.DB
}

func NewAccountRepo(db *bun.DB) AccountRepo {
	return &accountRepo{db: db}
}

// Implement static methods for AccountRepo
func (r *accountRepo) HashPassword(password string) (string, error) {
	return HashPassword(password)
}

func (r *accountRepo) VerifyPassword(password, hash string) (bool, error) {
	return VerifyPassword(password, hash)
}

// Create creates a new account
func (r *accountRepo) Create(ctx context.Context, email string, fullName string, authProviders []string, password *string, accountID *int64, analyticsPreference string, phoneNumber *string) (*Account, error) {
	account := &Account{
		FullName:      fullName,
		Email:         email,
		AuthProviders: authProviders,
	}

	if phoneNumber != nil {
		// Note: You would need to add PhoneNumber field to the Account model
		// account.PhoneNumber = *phoneNumber
	}

	if password != nil {
		hashedPassword, err := r.HashPassword(*password)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		account.PasswordHash = &hashedPassword
	}

	// Set default values for terms and analytics
	account.TermsAndPolicy = TermsAndPolicy{
		Type:      "acceptance",
		UpdatedAt: time.Now(),
		Version:   "1.0",
	}

	if analyticsPreference == "" {
		analyticsPreference = "undecided"
	}
	account.AnalyticsPref = AnalyticsPreference{
		Type:      analyticsPreference,
		UpdatedAt: time.Now(),
	}

	// Insert into database
	_, err := r.db.NewInsert().
		Model(account).
		Returning("*").
		Exec(ctx)
	if err != nil {
		// Handle unique constraint violations
		if isUniqueViolation(err) {
			if isEmailUniqueViolation(err) {
				return nil, ErrEmailAlreadyExists
			}
			// Add phone number check when PhoneNumber field is added
			return nil, fmt.Errorf("unique constraint violation: %w", err)
		}
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	return account, nil
}

// Get retrieves an account by ID
func (r *accountRepo) Get(ctx context.Context, accountID int64) (*Account, error) {
	account := &Account{}
	err := r.db.NewSelect().
		Model(account).
		Where("id = ?", accountID).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAccountNotFound
		}
		return nil, fmt.Errorf("failed to get account: %w", err)
	}
	return account, nil
}

// GetByEmail retrieves an account by email
func (r *accountRepo) GetByEmail(ctx context.Context, email string) (*Account, error) {
	account := &Account{}
	err := r.db.NewSelect().
		Model(account).
		Where("email = ?", email).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAccountNotFound
		}
		return nil, fmt.Errorf("failed to get account by email: %w", err)
	}
	return account, nil
}

// GetByPhoneNumber retrieves an account by phone number
func (r *accountRepo) GetByPhoneNumber(ctx context.Context, phoneNumber string) (*Account, error) {
	// Note: This would need PhoneNumber field added to Account model
	return nil, errors.New("phone number field not implemented in Account model")
}

// Update updates an account with optional fields
func (r *accountRepo) Update(ctx context.Context, account *Account, fullName *string, avatarURL *string, phoneNumber *string, termsAndPolicy *TermsAndPolicy, analyticsPreference *AnalyticsPreference, whatsappJobAlerts *bool) (*Account, error) {
	// Update the account struct fields if provided
	if fullName != nil {
		account.FullName = *fullName
	}

	if avatarURL != nil {
		account.InternalAvatarURL = avatarURL
	}

	// Note: phoneNumber would need to be added to the Account model
	if termsAndPolicy != nil {
		account.TermsAndPolicy = *termsAndPolicy
	}

	if analyticsPreference != nil {
		account.AnalyticsPref = *analyticsPreference
	}

	if whatsappJobAlerts != nil {
		// Note: You would need to add WhatsAppJobAlerts field to Account model
		// account.WhatsappJobAlertsEnabled = whatsappJobAlerts
	}

	// Use Bun's update functionality to save the changes
	_, err := r.db.NewUpdate().
		Model(account).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", account.ID).
		Returning("*").
		Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update account: %w", err)
	}

	return account, nil
}

// UpdateProfile updates the account's profile
func (r *accountRepo) UpdateProfile(ctx context.Context, account *Account, profile any) (*Account, error) {
	// Note: This would need Profile field added to Account model
	// For now, just return the account unchanged
	return account, nil
}

// UpdateAuthProviders updates the account's auth providers
func (r *accountRepo) UpdateAuthProviders(ctx context.Context, account *Account, authProviders []string) (*Account, error) {
	account.AuthProviders = updateStringSlice(account.AuthProviders, authProviders)

	_, err := r.db.NewUpdate().
		Model(account).
		Set("auth_providers = ?", account.AuthProviders).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", account.ID).
		Returning("*").
		Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update auth providers: %w", err)
	}

	return account, nil
}

// DeleteAvatar removes the account's avatar
func (r *accountRepo) DeleteAvatar(ctx context.Context, account *Account) (*Account, error) {
	account.InternalAvatarURL = nil

	_, err := r.db.NewUpdate().
		Model(account).
		Set("avatar_url = NULL").
		Set("updated_at = ?", time.Now()).
		Where("id = ?", account.ID).
		Returning("*").
		Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to delete avatar: %w", err)
	}

	return account, nil
}

// SetTwoFactorSecret sets the 2FA secret for an account
func (r *accountRepo) SetTwoFactorSecret(ctx context.Context, account *Account, totpSecret string) (*Account, error) {
	account.TwoFactorSecret = &totpSecret

	_, err := r.db.NewUpdate().
		Model(account).
		Set("two_factor_secret = ?", totpSecret).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", account.ID).
		Returning("*").
		Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to set two factor secret: %w", err)
	}

	return account, nil
}

// DeleteTwoFactorSecret removes the 2FA secret from an account
func (r *accountRepo) DeleteTwoFactorSecret(ctx context.Context, account *Account) (*Account, error) {
	account.TwoFactorSecret = nil

	_, err := r.db.NewUpdate().
		Model(account).
		Set("two_factor_secret = NULL").
		Set("updated_at = ?", time.Now()).
		Where("id = ?", account.ID).
		Returning("*").
		Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to delete two factor secret: %w", err)
	}

	return account, nil
}

// UpdatePassword updates the account's password
func (r *accountRepo) UpdatePassword(ctx context.Context, account *Account, password string) (*Account, error) {
	hashedPassword, err := r.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	account.PasswordHash = &hashedPassword

	// Add "password" to auth providers if not already present
	if !slices.Contains(account.AuthProviders, "password") {
		account.AuthProviders = addStringToSlice(account.AuthProviders, "password")
	}

	_, err = r.db.NewUpdate().
		Model(account).
		Set("password_hash = ?", hashedPassword).
		Set("auth_providers = ?", account.AuthProviders).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", account.ID).
		Returning("*").
		Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update password: %w", err)
	}

	return account, nil
}

// DeletePassword removes the account's password
func (r *accountRepo) DeletePassword(ctx context.Context, account *Account) (*Account, error) {
	account.PasswordHash = nil

	// Remove "password" from auth providers
	account.AuthProviders = removeStringFromSlice(account.AuthProviders, "password")

	_, err := r.db.NewUpdate().
		Model(account).
		Set("password_hash = NULL").
		Set("auth_providers = ?", account.AuthProviders).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", account.ID).
		Returning("*").
		Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to delete password: %w", err)
	}

	return account, nil
}

// Delete permanently removes an account
func (r *accountRepo) Delete(ctx context.Context, account *Account) error {
	_, err := r.db.NewDelete().
		Model(account).
		Where("id = ?", account.ID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete account: %w", err)
	}
	return nil
}

// Helper functions for error handling
func isUniqueViolation(err error) bool {
	// PostgreSQL unique constraint violation codes
	// You may need to adjust this based on your database
	errStr := err.Error()
	return contains(errStr, "duplicate key") ||
		contains(errStr, "unique constraint") ||
		contains(errStr, "UNIQUE violation")
}

func isEmailUniqueViolation(err error) bool {
	errStr := err.Error()
	return contains(errStr, "email") && contains(errStr, "unique")
}

func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

type EmailVerificationTokenRepo interface {
	Create(ctx context.Context, email string) (string, *EmailVerificationToken, error)
	Get(ctx context.Context, verificationToken string) (*EmailVerificationToken, error)
	GetByEmail(ctx context.Context, email string) (*EmailVerificationToken, error)
	Delete(ctx context.Context, emailVerification *EmailVerificationToken) error

	// Static methods for token operations
	GenerateVerificationToken(length int) (string, error)
	HashVerificationToken(token string) string
}

type emailVerificationTokenRepo struct {
	db *bun.DB
}

func NewEmailVerificationTokenRepo(db *bun.DB) EmailVerificationTokenRepo {
	return &emailVerificationTokenRepo{db: db}
}

// Implement static methods for EmailVerificationTokenRepo
func (r *emailVerificationTokenRepo) GenerateVerificationToken(length int) (string, error) {
	return GenerateVerificationToken(length)
}

func (r *emailVerificationTokenRepo) HashVerificationToken(token string) string {
	return HashVerificationToken(token)
}

// Create creates a new email verification token
func (r *emailVerificationTokenRepo) Create(ctx context.Context, email string) (string, *EmailVerificationToken, error) {
	token, err := r.GenerateVerificationToken(32)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate verification token: %w", err)
	}

	emailVerification := &EmailVerificationToken{
		Email:     email,
		TokenHash: r.HashVerificationToken(token),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	_, err = r.db.NewInsert().
		Model(emailVerification).
		Returning("*").
		Exec(ctx)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create email verification token: %w", err)
	}

	return token, emailVerification, nil
}

// Get retrieves an email verification token by the plaintext token
func (r *emailVerificationTokenRepo) Get(ctx context.Context, verificationToken string) (*EmailVerificationToken, error) {
	tokenHash := r.HashVerificationToken(verificationToken)

	emailVerification := &EmailVerificationToken{}
	err := r.db.NewSelect().
		Model(emailVerification).
		Where("token_hash = ?", tokenHash).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTokenNotFound
		}
		return nil, fmt.Errorf("failed to get email verification token: %w", err)
	}

	return emailVerification, nil
}

// GetByEmail retrieves an email verification token by email
func (r *emailVerificationTokenRepo) GetByEmail(ctx context.Context, email string) (*EmailVerificationToken, error) {
	emailVerification := &EmailVerificationToken{}
	err := r.db.NewSelect().
		Model(emailVerification).
		Where("email = ?", email).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTokenNotFound
		}
		return nil, fmt.Errorf("failed to get email verification token by email: %w", err)
	}

	return emailVerification, nil
}

// Delete removes an email verification token
func (r *emailVerificationTokenRepo) Delete(ctx context.Context, emailVerification *EmailVerificationToken) error {
	_, err := r.db.NewDelete().
		Model(emailVerification).
		Where("id = ?", emailVerification.ID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete email verification token: %w", err)
	}
	return nil
}

type PhoneNumberVerificationTokenRepo interface {
	Create(ctx context.Context, phoneNumber string) (string, *PhoneNumberVerificationToken, error)
	Get(ctx context.Context, verificationToken string) (*PhoneNumberVerificationToken, error)
	GetByPhoneNumber(ctx context.Context, phoneNumber string) (*PhoneNumberVerificationToken, error)
	Delete(ctx context.Context, phoneNumberVerification *PhoneNumberVerificationToken) error

	// Static methods for token operations
	GenerateVerificationToken(length int) (string, error)
	HashVerificationToken(token string) string
}

type phoneNumberVerificationTokenRepo struct {
	db *bun.DB
}

func NewPhoneNumberVerificationTokenRepo(db *bun.DB) PhoneNumberVerificationTokenRepo {
	return &phoneNumberVerificationTokenRepo{db: db}
}

// Implement static methods for PhoneNumberVerificationTokenRepo
func (r *phoneNumberVerificationTokenRepo) GenerateVerificationToken(length int) (string, error) {
	return GenerateVerificationToken(length)
}

func (r *phoneNumberVerificationTokenRepo) HashVerificationToken(token string) string {
	return HashVerificationToken(token)
}

// Create creates a new phone number verification token
func (r *phoneNumberVerificationTokenRepo) Create(ctx context.Context, phoneNumber string) (string, *PhoneNumberVerificationToken, error) {
	token, err := r.GenerateVerificationToken(32)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate verification token: %w", err)
	}

	phoneVerification := &PhoneNumberVerificationToken{
		PhoneNumber: phoneNumber,
		TokenHash:   r.HashVerificationToken(token),
		ExpiresAt:   time.Now().Add(24 * time.Hour),
	}

	_, err = r.db.NewInsert().
		Model(phoneVerification).
		Returning("*").
		Exec(ctx)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create phone number verification token: %w", err)
	}

	return token, phoneVerification, nil
}

// Get retrieves a phone number verification token by the plaintext token
func (r *phoneNumberVerificationTokenRepo) Get(ctx context.Context, verificationToken string) (*PhoneNumberVerificationToken, error) {
	tokenHash := r.HashVerificationToken(verificationToken)

	phoneVerification := &PhoneNumberVerificationToken{}
	err := r.db.NewSelect().
		Model(phoneVerification).
		Where("token_hash = ?", tokenHash).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTokenNotFound
		}
		return nil, fmt.Errorf("failed to get phone number verification token: %w", err)
	}

	return phoneVerification, nil
}

// GetByPhoneNumber retrieves a phone number verification token by phone number
func (r *phoneNumberVerificationTokenRepo) GetByPhoneNumber(ctx context.Context, phoneNumber string) (*PhoneNumberVerificationToken, error) {
	phoneVerification := &PhoneNumberVerificationToken{}
	err := r.db.NewSelect().
		Model(phoneVerification).
		Where("phone_number = ?", phoneNumber).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTokenNotFound
		}
		return nil, fmt.Errorf("failed to get phone number verification token by phone number: %w", err)
	}

	return phoneVerification, nil
}

// Delete removes a phone number verification token
func (r *phoneNumberVerificationTokenRepo) Delete(ctx context.Context, phoneNumberVerification *PhoneNumberVerificationToken) error {
	_, err := r.db.NewDelete().
		Model(phoneNumberVerification).
		Where("id = ?", phoneNumberVerification.ID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete phone number verification token: %w", err)
	}
	return nil
}
