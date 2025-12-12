package auth

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
	"database/sql"
)

// setupTestDB creates a test database connection for testing
func setupTestDB(t *testing.T) *bun.DB {
	// For unit tests, we'll skip actual DB connection since we're not testing database operations
	// In a real implementation, you might use testcontainers or an in-memory DB
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN("postgres://postgres:password@localhost:5432/test_db?sslmode=disable")))

	db := bun.NewDB(sqldb, pgdialect.New())

	// Add debug for testing
	db.AddQueryHook(bundebug.NewQueryHook(
		bundebug.WithVerbose(true),
	))

	return db
}

func TestAuthRepoInterfaceSatisfaction(t *testing.T) {
	t.Run("All repository interfaces are properly implemented", func(t *testing.T) {
		// This test ensures our implementations satisfy the interfaces
		var _ SessionRepo = (*sessionRepo)(nil)
		var _ PasswordResetTokenRepo = (*passwordResetTokenRepo)(nil)
		var _ WebAuthnCredentialRepo = (*webAuthnCredentialRepo)(nil)
		var _ WebAuthnChallengeRepo = (*webAuthnChallengeRepo)(nil)
		var _ OAuthCredentialRepo = (*oAuthCredentialRepo)(nil)
		var _ TwoFactorAuthenticationChallengeRepo = (*twoFactorAuthenticationChallengeRepo)(nil)
		var _ RecoveryCodeRepo = (*recoveryCodeRepo)(nil)
		var _ TemporaryTwoFactorChallengeRepo = (*temporaryTwoFactorChallengeRepo)(nil)

		// If we reach here, all implementations satisfy their interfaces
		assert.True(t, true, "All repository implementations satisfy their interfaces")
	})
}

func TestSessionRepoStaticMethods(t *testing.T) {
	repo := &sessionRepo{}

	t.Run("GenerateSessionToken creates unique tokens", func(t *testing.T) {
		token1, err := repo.GenerateSessionToken()
		require.NoError(t, err)
		assert.NotEmpty(t, token1)
		assert.Len(t, token1, 64) // 32 bytes = 64 hex chars

		token2, err := repo.GenerateSessionToken()
		require.NoError(t, err)
		assert.NotEmpty(t, token2)
		assert.NotEqual(t, token1, token2)
	})

	t.Run("HashSessionToken produces consistent hashes", func(t *testing.T) {
		token := "test-token-12345"
		hash1 := repo.HashSessionToken(token)
		hash2 := repo.HashSessionToken(token)

		assert.Equal(t, hash1, hash2)
		assert.NotEmpty(t, hash1)
		assert.Len(t, hash1, 32) // MD5 hash = 32 hex chars
	})
}

func TestPasswordResetTokenRepoStaticMethods(t *testing.T) {
	repo := &passwordResetTokenRepo{}

	t.Run("GeneratePasswordResetToken creates unique tokens", func(t *testing.T) {
		token1, err := repo.GeneratePasswordResetToken()
		require.NoError(t, err)
		assert.NotEmpty(t, token1)
		assert.Len(t, token1, 64)

		token2, err := repo.GeneratePasswordResetToken()
		require.NoError(t, err)
		assert.NotEqual(t, token1, token2)
	})

	t.Run("HashPasswordResetToken produces consistent hashes", func(t *testing.T) {
		token := "reset-token-12345"
		hash1 := repo.HashPasswordResetToken(token)
		hash2 := repo.HashPasswordResetToken(token)

		assert.Equal(t, hash1, hash2)
		assert.NotEmpty(t, hash1)
	})
}

func TestRecoveryCodeRepoStaticMethods(t *testing.T) {
	repo := &recoveryCodeRepo{}

	t.Run("GenerateRecoveryCode creates valid codes", func(t *testing.T) {
		code1, err := repo.GenerateRecoveryCode()
		require.NoError(t, err)
		assert.NotEmpty(t, code1)
		assert.Len(t, code1, 8)

		code2, err := repo.GenerateRecoveryCode()
		require.NoError(t, err)
		assert.NotEqual(t, code1, code2)
	})

	t.Run("HashRecoveryCode produces consistent hashes", func(t *testing.T) {
		code := "RECOVERY123"
		hash1 := repo.HashRecoveryCode(code)
		hash2 := repo.HashRecoveryCode(code)

		assert.Equal(t, hash1, hash2)
		assert.NotEmpty(t, hash1)
	})
}

func TestTwoFactorChallengeRepoStaticMethods(t *testing.T) {
	repo := &twoFactorAuthenticationChallengeRepo{}

	t.Run("GenerateChallenge creates unique challenges", func(t *testing.T) {
		challenge1, err := repo.GenerateChallenge()
		require.NoError(t, err)
		assert.NotEmpty(t, challenge1)
		assert.Len(t, challenge1, 64)

		challenge2, err := repo.GenerateChallenge()
		require.NoError(t, err)
		assert.NotEqual(t, challenge1, challenge2)
	})

	t.Run("HashChallenge produces consistent hashes", func(t *testing.T) {
		challenge := "challenge-12345"
		hash1 := repo.HashChallenge(challenge)
		hash2 := repo.HashChallenge(challenge)

		assert.Equal(t, hash1, hash2)
		assert.NotEmpty(t, hash1)
	})
}

func TestTemporaryTwoFactorChallengeRepoStaticMethods(t *testing.T) {
	repo := &temporaryTwoFactorChallengeRepo{}

	t.Run("GenerateChallenge creates unique challenges", func(t *testing.T) {
		challenge1, err := repo.GenerateChallenge()
		require.NoError(t, err)
		assert.NotEmpty(t, challenge1)

		challenge2, err := repo.GenerateChallenge()
		require.NoError(t, err)
		assert.NotEqual(t, challenge1, challenge2)
	})

	t.Run("HashChallenge produces consistent hashes", func(t *testing.T) {
		challenge := "temp-challenge-12345"
		hash1 := repo.HashChallenge(challenge)
		hash2 := repo.HashChallenge(challenge)

		assert.Equal(t, hash1, hash2)
		assert.NotEmpty(t, hash1)
	})
}

func TestErrorDefinitions(t *testing.T) {
	t.Run("All auth repository errors are defined", func(t *testing.T) {
		errors := []error{
			ErrSessionNotFound,
			ErrTokenExpired,
			ErrInvalidCredentials,
			ErrWebAuthnCredentialNotFound,
			ErrChallengeNotFound,
			ErrRecoveryCodeInvalid,
			ErrOAuthCredentialAlreadyExists,
			ErrPasswordResetTokenNotFound,
			ErrTwoFactorAuthenticationNotFound,
			ErrTemporaryTwoFactorNotFound,
		}

		for i, err := range errors {
			require.NotNil(t, err, "Error %d should be defined", i)
			assert.NotEmpty(t, err.Error(), "Error %d should have a message", i)
		}
	})
}

func TestUtilityFunctions(t *testing.T) {
	t.Run("generateSecureToken creates unique tokens", func(t *testing.T) {
		token1, err := generateSecureToken(16)
		require.NoError(t, err)
		assert.NotEmpty(t, token1)
		assert.Len(t, token1, 32) // 16 bytes = 32 hex chars

		token2, err := generateSecureToken(16)
		require.NoError(t, err)
		assert.NotEqual(t, token1, token2)
	})

	t.Run("hashTokenMD5 produces consistent hashes", func(t *testing.T) {
		input := "test-input"
		hash1 := hashTokenMD5(input)
		hash2 := hashTokenMD5(input)

		assert.Equal(t, hash1, hash2)
		assert.Len(t, hash1, 32) // MD5 hash = 32 hex chars
	})

	t.Run("isUniqueViolation detects unique constraint errors", func(t *testing.T) {
		// Test with PostgreSQL unique violation
		pgErr := fmt.Errorf("duplicate key value violates unique constraint")
		assert.True(t, isUniqueViolation(pgErr))

		// Test with non-unique error
		otherErr := fmt.Errorf("some other error")
		assert.False(t, isUniqueViolation(otherErr))
	})
}

func TestTwoFactorSecretGeneration(t *testing.T) {
	t.Run("generateTwoFactorSecret creates valid TOTP secrets", func(t *testing.T) {
		secret, err := generateTwoFactorSecret()
		require.NoError(t, err)
		assert.NotEmpty(t, secret)

		// TOTP secrets are typically base32 encoded
		// Length varies but should be reasonable
		assert.Greater(t, len(secret), 10)
		assert.Less(t, len(secret), 100)

		// Should be consistent for same input (though our function generates random)
		secret2, err := generateTwoFactorSecret()
		require.NoError(t, err)
		assert.NotEqual(t, secret, secret2)
	})
}

func TestRepositoryConstructors(t *testing.T) {
	// This tests that repository constructors create non-nil instances
	t.Run("All repository constructors work", func(t *testing.T) {
		// Create a mock DB that doesn't need actual connection for constructor test
		testDB := setupTestDB(t)

		sessionRepo := NewSessionRepo(testDB)
		assert.NotNil(t, sessionRepo)

		passwordResetRepo := NewPasswordResetTokenRepo(testDB)
		assert.NotNil(t, passwordResetRepo)

		webauthnCredRepo := NewWebAuthnCredentialRepo(testDB)
		assert.NotNil(t, webauthnCredRepo)

		webauthnChallengeRepo := NewWebAuthnChallengeRepo(testDB)
		assert.NotNil(t, webauthnChallengeRepo)

		oauthRepo := NewOAuthCredentialRepo(testDB)
		assert.NotNil(t, oauthRepo)

		twoFactorRepo := NewTwoFactorAuthenticationChallengeRepo(testDB)
		assert.NotNil(t, twoFactorRepo)

		recoveryRepo := NewRecoveryCodeRepo(testDB)
		assert.NotNil(t, recoveryRepo)

		tempTwoFactorRepo := NewTemporaryTwoFactorChallengeRepo(testDB)
		assert.NotNil(t, tempTwoFactorRepo)
	})
}

func TestTokenGenerationConsistency(t *testing.T) {
	t.Run("Token generation across repos produces different tokens", func(t *testing.T) {
		sessionRepo := &sessionRepo{}
		passwordResetRepo := &passwordResetTokenRepo{}
		twoFactorRepo := &twoFactorAuthenticationChallengeRepo{}
		tempTwoFactorRepo := &temporaryTwoFactorChallengeRepo{}

		sessionToken, _ := sessionRepo.GenerateSessionToken()
		resetToken, _ := passwordResetRepo.GeneratePasswordResetToken()
		challenge1, _ := twoFactorRepo.GenerateChallenge()
		challenge2, _ := tempTwoFactorRepo.GenerateChallenge()

		// All should be different (very high probability)
		assert.NotEqual(t, sessionToken, resetToken)
		assert.NotEqual(t, resetToken, challenge1)
		assert.NotEqual(t, challenge1, challenge2)
	})
}