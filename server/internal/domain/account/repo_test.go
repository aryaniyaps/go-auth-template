package account

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

// setupTestDB creates a test database connection
func setupTestDB(t *testing.T) *bun.DB {
	// Use in-memory PostgreSQL connection string for testing
	// In a real environment, you would use a test database
	connector := pgdriver.NewConnector(pgdriver.WithDSN("postgres://test:test@localhost:5432/testdb?sslmode=disable"))
	sqlDB := sql.OpenDB(connector)
	db := bun.NewDB(sqlDB, pgdialect.New())

	// Create tables
	ctx := context.Background()
	_, err := db.NewCreateTable().Model((*Account)(nil)).IfNotExists().Exec(ctx)
	if err != nil {
		t.Skipf("Skipping test due to database not being available: %v", err)
	}
	_, err = db.NewCreateTable().Model((*EmailVerificationToken)(nil)).IfNotExists().Exec(ctx)
	if err != nil {
		t.Skipf("Skipping test due to database not being available: %v", err)
	}
	_, err = db.NewCreateTable().Model((*PhoneNumberVerificationToken)(nil)).IfNotExists().Exec(ctx)
	if err != nil {
		t.Skipf("Skipping test due to database not being available: %v", err)
	}

	return db
}

// Task Group 1: Foundation and Security Utilities Tests

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "testPassword123!",
			wantErr:  false,
		},
		{
			name:     "short password",
			password: "abc",
			wantErr:  false,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  false, // Hashing should still work for empty passwords
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, hash)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, hash)
				// Hash should be base64 encoded
				assert.NotContains(t, hash, "\x00")
			}
		})
	}
}

func TestVerifyPassword(t *testing.T) {
	password := "testPassword123!"

	// First, hash a password
	hash, err := HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hash)

	tests := []struct {
		name         string
		password     string
		hash         string
		wantValid    bool
		wantErr      bool
	}{
		{
			name:      "correct password",
			password:  password,
			hash:      hash,
			wantValid: true,
			wantErr:   false,
		},
		{
			name:      "incorrect password",
			password:  "wrongPassword",
			hash:      hash,
			wantValid: false,
			wantErr:   false,
		},
		{
			name:      "empty password",
			password:  "",
			hash:      hash,
			wantValid: false,
			wantErr:   false,
		},
		{
			name:      "invalid hash format",
			password:  password,
			hash:      "invalid-hash",
			wantValid: false,
			wantErr:   true,
		},
		{
			name:      "empty hash",
			password:  password,
			hash:      "",
			wantValid: false,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := VerifyPassword(tt.password, tt.hash)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantValid, valid)
			}
		})
	}
}

func TestHashVerificationToken(t *testing.T) {
	token := "test-token-12345"

	hash1 := HashVerificationToken(token)
	hash2 := HashVerificationToken(token)

	// Same token should produce same hash
	assert.Equal(t, hash1, hash2)
	assert.NotEmpty(t, hash1)
	assert.Len(t, hash1, 32) // MD5 hash length in hex
}

func TestGenerateVerificationToken(t *testing.T) {
	tests := []struct {
		name    string
		length  int
		wantErr bool
	}{
		{
			name:    "valid length",
			length:  32,
			wantErr: false,
		},
		{
			name:    "zero length (should use default)",
			length:  0,
			wantErr: false,
		},
		{
			name:    "negative length (should use default)",
			length:  -1,
			wantErr: false,
		},
		{
			name:    "small length",
			length:  8,
			wantErr: false,
		},
		{
			name:    "large length",
			length:  64,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateVerificationToken(tt.length)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)

				// Expected length should be 2 * length (hex encoding of bytes)
				expectedLen := tt.length
				if expectedLen <= 0 {
					expectedLen = 32 // default
				}
				expectedLen *= 2 // hex encoding
				assert.Len(t, token, expectedLen)

				// Should be valid hex
				assert.Regexp(t, `^[0-9a-f]+$`, token)
			}
		})
	}
}

func TestGenerateVerificationTokenUniqueness(t *testing.T) {
	// Generate multiple tokens and ensure they're unique
	tokens := make([]string, 10)
	for i := range tokens {
		token, err := GenerateVerificationToken(16)
		require.NoError(t, err)
		tokens[i] = token
	}

	// Check uniqueness
	tokenSet := make(map[string]bool)
	for _, token := range tokens {
		assert.False(t, tokenSet[token], "Token should be unique: %s", token)
		tokenSet[token] = true
	}
}

// Task Group 2: Account Repository Basic CRUD Tests

func TestNewAccountRepo(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAccountRepo(db)

	assert.NotNil(t, repo)
	assert.Implements(t, (*AccountRepo)(nil), repo)
}

func TestAccountRepo_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAccountRepo(db)
	ctx := context.Background()

	tests := []struct {
		name                 string
		email                string
		fullName             string
		authProviders        []string
		password             *string
		analyticsPreference  string
		phoneNumber          *string
		wantErr              bool
		expectedErrContains  string
	}{
		{
			name:                "create account with password",
			email:               "test@example.com",
			fullName:            "Test User",
			authProviders:       []string{"password"},
			password:            stringPtr("password123"),
			analyticsPreference: "undecided",
			wantErr:             false,
		},
		{
			name:                "create account without password",
			email:               "oauth@example.com",
			fullName:            "OAuth User",
			authProviders:       []string{"google"},
			password:            nil,
			analyticsPreference: "acceptance",
			wantErr:             false,
		},
		{
			name:                "create account with multiple auth providers",
			email:               "multi@example.com",
			fullName:            "Multi Auth User",
			authProviders:       []string{"password", "google"},
			password:            stringPtr("password123"),
			analyticsPreference: "rejection",
			wantErr:             false,
		},
		{
			name:                "empty analytics preference should default",
			email:               "default@example.com",
			fullName:            "Default User",
			authProviders:       []string{"password"},
			password:            stringPtr("password123"),
			analyticsPreference: "",
			wantErr:             false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account, err := repo.Create(ctx, tt.email, tt.fullName, tt.authProviders, tt.password, nil, tt.analyticsPreference, tt.phoneNumber)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErrContains != "" {
					assert.Contains(t, err.Error(), tt.expectedErrContains)
				}
				assert.Nil(t, account)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, account)
				assert.NotZero(t, account.ID)
				assert.Equal(t, tt.email, account.Email)
				assert.Equal(t, tt.fullName, account.FullName)
				assert.Equal(t, tt.authProviders, account.AuthProviders)

				if tt.password != nil {
					assert.NotNil(t, account.PasswordHash)
					// Verify the password can be verified
					valid, err := repo.VerifyPassword(*tt.password, *account.PasswordHash)
					assert.NoError(t, err)
					assert.True(t, valid)
				} else {
					assert.Nil(t, account.PasswordHash)
				}

				// Check analytics preference
				expectedAnalytics := tt.analyticsPreference
				if expectedAnalytics == "" {
					expectedAnalytics = "undecided"
				}
				assert.Equal(t, expectedAnalytics, account.AnalyticsPref.Type)
			}
		})
	}
}

func TestAccountRepo_Get(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAccountRepo(db)
	ctx := context.Background()

	// Create an account first
	account, err := repo.Create(ctx, "get@example.com", "Get User", []string{"password"}, stringPtr("password123"), nil, "", nil)
	require.NoError(t, err)
	require.NotNil(t, account)

	tests := []struct {
		name       string
		accountID  int64
		wantErr    bool
		wantResult bool
	}{
		{
			name:       "get existing account",
			accountID:  account.ID,
			wantErr:    false,
			wantResult: true,
		},
		{
			name:       "get non-existent account",
			accountID:  99999,
			wantErr:    true,
			wantResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.Get(ctx, tt.accountID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrAccountNotFound)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				if tt.wantResult {
					assert.NotNil(t, result)
					assert.Equal(t, tt.accountID, result.ID)
				}
			}
		})
	}
}

func TestAccountRepo_GetByEmail(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAccountRepo(db)
	ctx := context.Background()

	// Create an account first
	account, err := repo.Create(ctx, "email@example.com", "Email User", []string{"password"}, stringPtr("password123"), nil, "", nil)
	require.NoError(t, err)
	require.NotNil(t, account)

	tests := []struct {
		name        string
		email       string
		wantErr     bool
		wantResult  bool
		expectedID  int64
	}{
		{
			name:       "get existing account by email",
			email:      "email@example.com",
			wantErr:    false,
			wantResult: true,
			expectedID: account.ID,
		},
		{
			name:       "get non-existent account by email",
			email:      "nonexistent@example.com",
			wantErr:    true,
			wantResult: false,
			expectedID: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetByEmail(ctx, tt.email)

			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrAccountNotFound)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				if tt.wantResult {
					assert.NotNil(t, result)
					assert.Equal(t, tt.expectedID, result.ID)
					assert.Equal(t, tt.email, result.Email)
				}
			}
		})
	}
}

// Task Group 3: Account Repository Update Operations Tests

func TestAccountRepo_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAccountRepo(db)
	ctx := context.Background()

	// Create an account first
	account, err := repo.Create(ctx, "update@example.com", "Update User", []string{"password"}, stringPtr("password123"), nil, "", nil)
	require.NoError(t, err)
	require.NotNil(t, account)

	originalName := account.FullName
	originalUpdatedAt := account.UpdatedAt

	// Wait a bit to ensure updated_at changes
	time.Sleep(10 * time.Millisecond)

	tests := []struct {
		name                  string
		fullName              *string
		avatarURL             *string
		phoneNumber           *string
		termsAndPolicy        *TermsAndPolicy
		analyticsPreference   *AnalyticsPreference
		whatsappJobAlerts     *bool
		expectNameChange      bool
		expectAvatarChange    bool
		expectPolicyChange    bool
		expectAnalyticsChange bool
	}{
		{
			name:                "update full name only",
			fullName:            stringPtr("Updated Name"),
			expectNameChange:    true,
			expectAvatarChange:  false,
		},
		{
			name:                "update avatar URL only",
			avatarURL:           stringPtr("https://example.com/avatar.jpg"),
			expectNameChange:    false,
			expectAvatarChange:  true,
		},
		{
			name:               "update terms and policy only",
			termsAndPolicy: &TermsAndPolicy{
				Type:      "updated",
				UpdatedAt: time.Now(),
				Version:   "2.0",
			},
			expectNameChange:     false,
			expectAvatarChange:   false,
			expectPolicyChange:   true,
		},
		{
			name:                "update analytics preference only",
			analyticsPreference: &AnalyticsPreference{
				Type:      "enabled",
				UpdatedAt: time.Now(),
			},
			expectNameChange:      false,
			expectAvatarChange:    false,
			expectAnalyticsChange: true,
		},
		{
			name:                "no updates",
			expectNameChange:    false,
			expectAvatarChange:  false,
			expectPolicyChange:  false,
			expectAnalyticsChange: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Get fresh account for each test
			freshAccount, err := repo.Get(ctx, account.ID)
			require.NoError(t, err)

			originalAvatarURL := freshAccount.InternalAvatarURL
			originalTerms := freshAccount.TermsAndPolicy
			originalAnalytics := freshAccount.AnalyticsPref

			updatedAccount, err := repo.Update(ctx, freshAccount, tt.fullName, tt.avatarURL, tt.phoneNumber, tt.termsAndPolicy, tt.analyticsPreference, tt.whatsappJobAlerts)

			assert.NoError(t, err)
			assert.NotNil(t, updatedAccount)

			// Updated timestamp should be newer than original
			assert.True(t, updatedAccount.UpdatedAt.After(originalUpdatedAt))

			// Check full name
			if tt.expectNameChange {
				assert.Equal(t, *tt.fullName, updatedAccount.FullName)
				assert.NotEqual(t, originalName, updatedAccount.FullName)
			} else {
				assert.Equal(t, originalName, updatedAccount.FullName)
			}

			// Check avatar URL
			if tt.expectAvatarChange {
				assert.Equal(t, tt.avatarURL, updatedAccount.InternalAvatarURL)
				assert.NotEqual(t, originalAvatarURL, updatedAccount.InternalAvatarURL)
			} else {
				assert.Equal(t, originalAvatarURL, updatedAccount.InternalAvatarURL)
			}

			// Check terms and policy
			if tt.expectPolicyChange {
				assert.Equal(t, tt.termsAndPolicy.Type, updatedAccount.TermsAndPolicy.Type)
				assert.NotEqual(t, originalTerms.Type, updatedAccount.TermsAndPolicy.Type)
			} else {
				assert.Equal(t, originalTerms.Type, updatedAccount.TermsAndPolicy.Type)
			}

			// Check analytics preference
			if tt.expectAnalyticsChange {
				assert.Equal(t, tt.analyticsPreference.Type, updatedAccount.AnalyticsPref.Type)
				assert.NotEqual(t, originalAnalytics.Type, updatedAccount.AnalyticsPref.Type)
			} else {
				assert.Equal(t, originalAnalytics.Type, updatedAccount.AnalyticsPref.Type)
			}
		})
	}
}

func TestAccountRepo_UpdateAuthProviders(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAccountRepo(db)
	ctx := context.Background()

	// Create an account first
	account, err := repo.Create(ctx, "auth@example.com", "Auth User", []string{"password"}, stringPtr("password123"), nil, "", nil)
	require.NoError(t, err)
	require.NotNil(t, account)

	tests := []struct {
		name              string
		newAuthProviders  []string
		expectedProviders []string
	}{
		{
			name:              "add google auth provider",
			newAuthProviders:  []string{"password", "google"},
			expectedProviders: []string{"password", "google"},
		},
		{
			name:              "replace with only oauth providers",
			newAuthProviders:  []string{"google", "github"},
			expectedProviders: []string{"google", "github"},
		},
		{
			name:              "empty auth providers",
			newAuthProviders:  []string{},
			expectedProviders: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updatedAccount, err := repo.UpdateAuthProviders(ctx, account, tt.newAuthProviders)

			assert.NoError(t, err)
			assert.NotNil(t, updatedAccount)
			assert.Equal(t, tt.expectedProviders, updatedAccount.AuthProviders)

			// Update the account for next test
			account = updatedAccount
		})
	}
}

func TestAccountRepo_DeleteAvatar(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAccountRepo(db)
	ctx := context.Background()

	// Create an account first
	account, err := repo.Create(ctx, "avatar@example.com", "Avatar User", []string{"password"}, stringPtr("password123"), nil, "", nil)
	require.NoError(t, err)
	require.NotNil(t, account)

	// Set an avatar URL first
	avatarURL := "https://example.com/avatar.jpg"
	account.InternalAvatarURL = &avatarURL
	_, err = db.NewUpdate().
		Model(account).
		Set("avatar_url = ?", avatarURL).
		Where("id = ?", account.ID).
		Exec(ctx)
	require.NoError(t, err)

	// Verify avatar is set
	accountWithAvatar, err := repo.Get(ctx, account.ID)
	require.NoError(t, err)
	assert.Equal(t, &avatarURL, accountWithAvatar.InternalAvatarURL)

	// Delete avatar
	updatedAccount, err := repo.DeleteAvatar(ctx, accountWithAvatar)
	assert.NoError(t, err)
	assert.NotNil(t, updatedAccount)
	assert.Nil(t, updatedAccount.InternalAvatarURL)

	// Verify avatar is deleted in database
	accountAfterDeletion, err := repo.Get(ctx, account.ID)
	require.NoError(t, err)
	assert.Nil(t, accountAfterDeletion.InternalAvatarURL)
}

// Task Group 4: Account Repository Security Operations Tests

func TestAccountRepo_UpdatePassword(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAccountRepo(db)
	ctx := context.Background()

	// Create an account first
	account, err := repo.Create(ctx, "password@example.com", "Password User", []string{"google"}, nil, nil, "", nil)
	require.NoError(t, err)
	require.NotNil(t, account)
	assert.Nil(t, account.PasswordHash)
	assert.NotContains(t, account.AuthProviders, "password")

	newPassword := "newPassword123!"
	updatedAccount, err := repo.UpdatePassword(ctx, account, newPassword)

	assert.NoError(t, err)
	assert.NotNil(t, updatedAccount)
	assert.NotNil(t, updatedAccount.PasswordHash)

	// Verify password can be verified
	valid, err := repo.VerifyPassword(newPassword, *updatedAccount.PasswordHash)
	assert.NoError(t, err)
	assert.True(t, valid)

	// Verify "password" was added to auth providers
	assert.Contains(t, updatedAccount.AuthProviders, "password")
}

func TestAccountRepo_UpdatePassword_AlreadyHasPassword(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAccountRepo(db)
	ctx := context.Background()

	// Create an account with password
	account, err := repo.Create(ctx, "existing@example.com", "Existing User", []string{"password"}, stringPtr("oldPassword123"), nil, "", nil)
	require.NoError(t, err)
	require.NotNil(t, account)

	originalAuthProvidersLen := len(account.AuthProviders)
	assert.Contains(t, account.AuthProviders, "password")

	newPassword := "newPassword456!"
	updatedAccount, err := repo.UpdatePassword(ctx, account, newPassword)

	assert.NoError(t, err)
	assert.NotNil(t, updatedAccount)

	// Verify new password works
	valid, err := repo.VerifyPassword(newPassword, *updatedAccount.PasswordHash)
	assert.NoError(t, err)
	assert.True(t, valid)

	// Verify old password doesn't work
	invalid, err := repo.VerifyPassword("oldPassword123", *updatedAccount.PasswordHash)
	assert.NoError(t, err)
	assert.False(t, invalid)

	// Verify "password" is still in auth providers (not duplicated)
	assert.Contains(t, updatedAccount.AuthProviders, "password")
	assert.Equal(t, originalAuthProvidersLen, len(updatedAccount.AuthProviders))
}

func TestAccountRepo_DeletePassword(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAccountRepo(db)
	ctx := context.Background()

	// Create an account with password
	account, err := repo.Create(ctx, "delete@example.com", "Delete User", []string{"password", "google"}, stringPtr("password123"), nil, "", nil)
	require.NoError(t, err)
	require.NotNil(t, account)
	assert.NotNil(t, account.PasswordHash)
	assert.Contains(t, account.AuthProviders, "password")

	updatedAccount, err := repo.DeletePassword(ctx, account)

	assert.NoError(t, err)
	assert.NotNil(t, updatedAccount)
	assert.Nil(t, updatedAccount.PasswordHash)
	assert.NotContains(t, updatedAccount.AuthProviders, "password")
	assert.Contains(t, updatedAccount.AuthProviders, "google") // Other auth providers should remain
}

func TestAccountRepo_SetTwoFactorSecret(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAccountRepo(db)
	ctx := context.Background()

	// Create an account first
	account, err := repo.Create(ctx, "2fa@example.com", "2FA User", []string{"password"}, stringPtr("password123"), nil, "", nil)
	require.NoError(t, err)
	require.NotNil(t, account)
	assert.Nil(t, account.TwoFactorSecret)

	secret := "JBSWY3DPEHPK3PXP"
	updatedAccount, err := repo.SetTwoFactorSecret(ctx, account, secret)

	assert.NoError(t, err)
	assert.NotNil(t, updatedAccount)
	assert.Equal(t, &secret, updatedAccount.TwoFactorSecret)

	// Verify 2FA is enabled
	assert.True(t, updatedAccount.Has2FAEnabled())
	providers := updatedAccount.TwoFactorProviders()
	assert.Len(t, providers, 1)
	assert.Equal(t, TwoFactorProviderAuthenticator, providers[0])
}

func TestAccountRepo_DeleteTwoFactorSecret(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAccountRepo(db)
	ctx := context.Background()

	// Create an account and set 2FA secret
	account, err := repo.Create(ctx, "n2fa@example.com", "No 2FA User", []string{"password"}, stringPtr("password123"), nil, "", nil)
	require.NoError(t, err)
	require.NotNil(t, account)

	secret := "JBSWY3DPEHPK3PXP"
	accountWith2FA, err := repo.SetTwoFactorSecret(ctx, account, secret)
	require.NoError(t, err)
	require.NotNil(t, accountWith2FA)
	assert.NotNil(t, accountWith2FA.TwoFactorSecret)

	// Delete 2FA secret
	updatedAccount, err := repo.DeleteTwoFactorSecret(ctx, accountWith2FA)

	assert.NoError(t, err)
	assert.NotNil(t, updatedAccount)
	assert.Nil(t, updatedAccount.TwoFactorSecret)

	// Verify 2FA is disabled
	assert.False(t, updatedAccount.Has2FAEnabled())
	providers := updatedAccount.TwoFactorProviders()
	assert.Len(t, providers, 0)
}

func TestAccountRepo_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAccountRepo(db)
	ctx := context.Background()

	// Create an account first
	account, err := repo.Create(ctx, "deleteuser@example.com", "Delete Me", []string{"password"}, stringPtr("password123"), nil, "", nil)
	require.NoError(t, err)
	require.NotNil(t, account)

	accountID := account.ID

	// Verify account exists
	existingAccount, err := repo.Get(ctx, accountID)
	require.NoError(t, err)
	require.NotNil(t, existingAccount)

	// Delete account
	err = repo.Delete(ctx, account)
	assert.NoError(t, err)

	// Verify account is deleted
	deletedAccount, err := repo.Get(ctx, accountID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrAccountNotFound)
	assert.Nil(t, deletedAccount)
}

func TestAccountRepo_StaticMethods(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAccountRepo(db)

	password := "testPassword123!"

	// Test static HashPassword method
	hash, err := repo.HashPassword(password)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)

	// Test static VerifyPassword method
	valid, err := repo.VerifyPassword(password, hash)
	assert.NoError(t, err)
	assert.True(t, valid)

	invalid, err := repo.VerifyPassword("wrongPassword", hash)
	assert.NoError(t, err)
	assert.False(t, invalid)
}

// Task Group 5: Email Verification Token Repository Tests

func TestNewEmailVerificationTokenRepo(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEmailVerificationTokenRepo(db)

	assert.NotNil(t, repo)
	assert.Implements(t, (*EmailVerificationTokenRepo)(nil), repo)
}

func TestEmailVerificationTokenRepo_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEmailVerificationTokenRepo(db)
	ctx := context.Background()

	email := "verify@example.com"
	token, emailToken, err := repo.Create(ctx, email)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.NotNil(t, emailToken)
	assert.NotZero(t, emailToken.ID)
	assert.Equal(t, email, emailToken.Email)
	assert.NotEmpty(t, emailToken.TokenHash)
	assert.True(t, time.Now().Add(23*time.Hour).Before(emailToken.ExpiresAt)) // Should expire in ~24 hours
	assert.True(t, time.Now().Add(25*time.Hour).After(emailToken.ExpiresAt))

	// Token should be cryptographically secure and hex encoded
	assert.Regexp(t, `^[0-9a-f]+$`, token)
	assert.Len(t, token, 64) // 32 bytes = 64 hex chars

	// Token hash should match MD5 hash of token
	expectedHash := HashVerificationToken(token)
	assert.Equal(t, expectedHash, emailToken.TokenHash)
}

func TestEmailVerificationTokenRepo_Get(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEmailVerificationTokenRepo(db)
	ctx := context.Background()

	// Create a token first
	email := "get@example.com"
	plaintextToken, emailToken, err := repo.Create(ctx, email)
	require.NoError(t, err)
	require.NotNil(t, emailToken)

	// Get token using plaintext token
	retrievedToken, err := repo.Get(ctx, plaintextToken)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedToken)
	assert.Equal(t, emailToken.ID, retrievedToken.ID)
	assert.Equal(t, email, retrievedToken.Email)
	assert.Equal(t, emailToken.TokenHash, retrievedToken.TokenHash)

	// Test with non-existent token
	nonExistentToken, err := repo.Get(ctx, "non-existent-token")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrTokenNotFound)
	assert.Nil(t, nonExistentToken)
}

func TestEmailVerificationTokenRepo_GetByEmail(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEmailVerificationTokenRepo(db)
	ctx := context.Background()

	// Create a token first
	email := "getbyemail@example.com"
	_, emailToken, err := repo.Create(ctx, email)
	require.NoError(t, err)
	require.NotNil(t, emailToken)

	// Get token by email
	retrievedToken, err := repo.GetByEmail(ctx, email)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedToken)
	assert.Equal(t, emailToken.ID, retrievedToken.ID)
	assert.Equal(t, email, retrievedToken.Email)

	// Test with non-existent email
	nonExistentToken, err := repo.GetByEmail(ctx, "non-existent@example.com")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrTokenNotFound)
	assert.Nil(t, nonExistentToken)
}

func TestEmailVerificationTokenRepo_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEmailVerificationTokenRepo(db)
	ctx := context.Background()

	// Create a token first
	email := "delete@example.com"
	_, emailToken, err := repo.Create(ctx, email)
	require.NoError(t, err)
	require.NotNil(t, emailToken)

	tokenID := emailToken.ID

	// Verify token exists before deletion
	existingToken, err := repo.GetByEmail(ctx, email)
	require.NoError(t, err)
	require.NotNil(t, existingToken)
	assert.Equal(t, tokenID, existingToken.ID)

	// Delete token
	err = repo.Delete(ctx, emailToken)
	assert.NoError(t, err)

	// Verify token is deleted
	deletedToken, err := repo.GetByEmail(ctx, email)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrTokenNotFound)
	assert.Nil(t, deletedToken)
}

func TestEmailVerificationTokenRepo_StaticMethods(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEmailVerificationTokenRepo(db)

	// Test static GenerateVerificationToken method
	token, err := repo.GenerateVerificationToken(16)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Len(t, token, 32) // 16 bytes = 32 hex chars
	assert.Regexp(t, `^[0-9a-f]+$`, token)

	// Test static HashVerificationToken method
	originalToken := "test-token-12345"
	hash1 := repo.HashVerificationToken(originalToken)
	hash2 := HashVerificationToken(originalToken) // Compare with global function

	assert.Equal(t, hash1, hash2)
	assert.Len(t, hash1, 32) // MD5 hash length in hex
}

// Task Group 6: Phone Number Verification Token Repository Tests

func TestNewPhoneNumberVerificationTokenRepo(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPhoneNumberVerificationTokenRepo(db)

	assert.NotNil(t, repo)
	assert.Implements(t, (*PhoneNumberVerificationTokenRepo)(nil), repo)
}

func TestPhoneNumberVerificationTokenRepo_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPhoneNumberVerificationTokenRepo(db)
	ctx := context.Background()

	phoneNumber := "+1234567890"
	token, phoneToken, err := repo.Create(ctx, phoneNumber)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.NotNil(t, phoneToken)
	assert.NotZero(t, phoneToken.ID)
	assert.Equal(t, phoneNumber, phoneToken.PhoneNumber)
	assert.NotEmpty(t, phoneToken.TokenHash)
	assert.True(t, time.Now().Add(23*time.Hour).Before(phoneToken.ExpiresAt)) // Should expire in ~24 hours
	assert.True(t, time.Now().Add(25*time.Hour).After(phoneToken.ExpiresAt))

	// Token should be cryptographically secure and hex encoded
	assert.Regexp(t, `^[0-9a-f]+$`, token)
	assert.Len(t, token, 64) // 32 bytes = 64 hex chars

	// Token hash should match MD5 hash of token
	expectedHash := HashVerificationToken(token)
	assert.Equal(t, expectedHash, phoneToken.TokenHash)
}

func TestPhoneNumberVerificationTokenRepo_Get(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPhoneNumberVerificationTokenRepo(db)
	ctx := context.Background()

	// Create a token first
	phoneNumber := "+1234567890"
	plaintextToken, phoneToken, err := repo.Create(ctx, phoneNumber)
	require.NoError(t, err)
	require.NotNil(t, phoneToken)

	// Get token using plaintext token
	retrievedToken, err := repo.Get(ctx, plaintextToken)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedToken)
	assert.Equal(t, phoneToken.ID, retrievedToken.ID)
	assert.Equal(t, phoneNumber, retrievedToken.PhoneNumber)
	assert.Equal(t, phoneToken.TokenHash, retrievedToken.TokenHash)

	// Test with non-existent token
	nonExistentToken, err := repo.Get(ctx, "non-existent-token")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrTokenNotFound)
	assert.Nil(t, nonExistentToken)
}

func TestPhoneNumberVerificationTokenRepo_GetByPhoneNumber(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPhoneNumberVerificationTokenRepo(db)
	ctx := context.Background()

	// Create a token first
	phoneNumber := "+1987654321"
	_, phoneToken, err := repo.Create(ctx, phoneNumber)
	require.NoError(t, err)
	require.NotNil(t, phoneToken)

	// Get token by phone number
	retrievedToken, err := repo.GetByPhoneNumber(ctx, phoneNumber)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedToken)
	assert.Equal(t, phoneToken.ID, retrievedToken.ID)
	assert.Equal(t, phoneNumber, retrievedToken.PhoneNumber)

	// Test with non-existent phone number
	nonExistentToken, err := repo.GetByPhoneNumber(ctx, "+1111111111")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrTokenNotFound)
	assert.Nil(t, nonExistentToken)
}

func TestPhoneNumberVerificationTokenRepo_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPhoneNumberVerificationTokenRepo(db)
	ctx := context.Background()

	// Create a token first
	phoneNumber := "+1555555555"
	_, phoneToken, err := repo.Create(ctx, phoneNumber)
	require.NoError(t, err)
	require.NotNil(t, phoneToken)

	tokenID := phoneToken.ID

	// Verify token exists before deletion
	existingToken, err := repo.GetByPhoneNumber(ctx, phoneNumber)
	require.NoError(t, err)
	require.NotNil(t, existingToken)
	assert.Equal(t, tokenID, existingToken.ID)

	// Delete token
	err = repo.Delete(ctx, phoneToken)
	assert.NoError(t, err)

	// Verify token is deleted
	deletedToken, err := repo.GetByPhoneNumber(ctx, phoneNumber)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrTokenNotFound)
	assert.Nil(t, deletedToken)
}

func TestPhoneNumberVerificationTokenRepo_StaticMethods(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPhoneNumberVerificationTokenRepo(db)

	// Test static GenerateVerificationToken method
	token, err := repo.GenerateVerificationToken(24)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Len(t, token, 48) // 24 bytes = 48 hex chars
	assert.Regexp(t, `^[0-9a-f]+$`, token)

	// Test static HashVerificationToken method
	originalToken := "phone-token-67890"
	hash1 := repo.HashVerificationToken(originalToken)
	hash2 := HashVerificationToken(originalToken) // Compare with global function

	assert.Equal(t, hash1, hash2)
	assert.Len(t, hash1, 32) // MD5 hash length in hex
}

// Task Group 7: Integration Tests

func TestRepositoryIntegration_CreateAccountAndVerificationToken(t *testing.T) {
	db := setupTestDB(t)
	accountRepo := NewAccountRepo(db)
	emailTokenRepo := NewEmailVerificationTokenRepo(db)
	phoneTokenRepo := NewPhoneNumberVerificationTokenRepo(db)
	ctx := context.Background()

	// Create an account
	email := "integration@example.com"
	password := "integrationPassword123!"
	account, err := accountRepo.Create(ctx, email, "Integration User", []string{"password"}, &password, nil, "", nil)
	require.NoError(t, err)
	require.NotNil(t, account)

	// Create email verification token
	emailToken, emailVerification, err := emailTokenRepo.Create(ctx, email)
	require.NoError(t, err)
	require.NotEmpty(t, emailToken)
	require.NotNil(t, emailVerification)

	// Create phone verification token
	phoneNumber := "+1234567890"
	phoneToken, phoneVerification, err := phoneTokenRepo.Create(ctx, phoneNumber)
	require.NoError(t, err)
	require.NotEmpty(t, phoneToken)
	require.NotNil(t, phoneVerification)

	// Verify we can retrieve all entities
	retrievedAccount, err := accountRepo.GetByEmail(ctx, email)
	require.NoError(t, err)
	require.Equal(t, account.ID, retrievedAccount.ID)

	retrievedEmailToken, err := emailTokenRepo.Get(ctx, emailToken)
	require.NoError(t, err)
	require.Equal(t, emailVerification.ID, retrievedEmailToken.ID)

	retrievedPhoneToken, err := phoneTokenRepo.Get(ctx, phoneToken)
	require.NoError(t, err)
	require.Equal(t, phoneVerification.ID, retrievedPhoneToken.ID)

	// Clean up
	err = emailTokenRepo.Delete(ctx, retrievedEmailToken)
	assert.NoError(t, err)
	err = phoneTokenRepo.Delete(ctx, retrievedPhoneToken)
	assert.NoError(t, err)
	err = accountRepo.Delete(ctx, retrievedAccount)
	assert.NoError(t, err)
}

func TestRepositoryInterface_Compliance(t *testing.T) {
	db := setupTestDB(t)

	// Test that all repositories implement their interfaces correctly
	var _ AccountRepo = NewAccountRepo(db)
	var _ EmailVerificationTokenRepo = NewEmailVerificationTokenRepo(db)
	var _ PhoneNumberVerificationTokenRepo = NewPhoneNumberVerificationTokenRepo(db)

	// Test that static methods work
	accountRepo := NewAccountRepo(db)
	emailRepo := NewEmailVerificationTokenRepo(db)
	phoneRepo := NewPhoneNumberVerificationTokenRepo(db)

	// Test password methods
	hash, err := accountRepo.HashPassword("test")
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)

	valid, err := accountRepo.VerifyPassword("test", hash)
	assert.NoError(t, err)
	assert.True(t, valid)

	// Test token methods
	token1, err := emailRepo.GenerateVerificationToken(32)
	assert.NoError(t, err)
	assert.NotEmpty(t, token1)

	token2, err := phoneRepo.GenerateVerificationToken(32)
	assert.NoError(t, err)
	assert.NotEmpty(t, token2)

	hash1 := emailRepo.HashVerificationToken("test")
	hash2 := phoneRepo.HashVerificationToken("test")
	assert.Equal(t, hash1, hash2)
	assert.Equal(t, HashVerificationToken("test"), hash1)
}

// Helper function
func stringPtr(s string) *string {
	return &s
}