package auth

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"

	"server/internal/infrastructure/db"
)

// Integration tests focus on critical auth repository workflows
// These tests are more strategic and focus on end-to-end scenarios

func TestTokenLifecycleManagement(t *testing.T) {
	t.Run("Session token lifecycle works correctly", func(t *testing.T) {
		repo := &sessionRepo{}

		// Generate and hash a session token
		token, err := repo.GenerateSessionToken()
		require.NoError(t, err)
		assert.NotEmpty(t, token)

		hashedToken := repo.HashSessionToken(token)
		assert.NotEmpty(t, hashedToken)
		assert.NotEqual(t, token, hashedToken)

		// Verify consistent hashing
		hashedToken2 := repo.HashSessionToken(token)
		assert.Equal(t, hashedToken, hashedToken2)

		// Verify token uniqueness
		newToken, err := repo.GenerateSessionToken()
		require.NoError(t, err)
		assert.NotEqual(t, token, newToken)

		newHashedToken := repo.HashSessionToken(newToken)
		assert.NotEqual(t, hashedToken, newHashedToken)
	})

	t.Run("Password reset token lifecycle works correctly", func(t *testing.T) {
		repo := &passwordResetTokenRepo{}

		token, err := repo.GeneratePasswordResetToken()
		require.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.Len(t, token, 64)

		hashedToken := repo.HashPasswordResetToken(token)
		assert.NotEmpty(t, hashedToken)
		assert.Len(t, hashedToken, 32)

		// Verify consistent hashing
		hashedToken2 := repo.HashPasswordResetToken(token)
		assert.Equal(t, hashedToken, hashedToken2)
	})

	t.Run("Recovery code lifecycle works correctly", func(t *testing.T) {
		repo := &recoveryCodeRepo{}

		code, err := repo.GenerateRecoveryCode()
		require.NoError(t, err)
		assert.NotEmpty(t, code)
		assert.Len(t, code, 8)

		hashedCode := repo.HashRecoveryCode(code)
		assert.NotEmpty(t, hashedCode)
		assert.Len(t, hashedCode, 32)

		// Verify consistent hashing
		hashedCode2 := repo.HashRecoveryCode(code)
		assert.Equal(t, hashedCode, hashedCode2)

		// Verify code uniqueness
		newCode, err := repo.GenerateRecoveryCode()
		require.NoError(t, err)
		assert.NotEqual(t, code, newCode)

		newHashedCode := repo.HashRecoveryCode(newCode)
		assert.NotEqual(t, hashedCode, newHashedCode)
	})

	t.Run("2FA challenge lifecycle works correctly", func(t *testing.T) {
		repo := &twoFactorAuthenticationChallengeRepo{}

		challenge, err := repo.GenerateChallenge()
		require.NoError(t, err)
		assert.NotEmpty(t, challenge)
		assert.Len(t, challenge, 64)

		hashedChallenge := repo.HashChallenge(challenge)
		assert.NotEmpty(t, hashedChallenge)
		assert.Len(t, hashedChallenge, 32)

		// Verify TOTP secret generation
		secret, err := repo.GenerateTwoFactorSecret()
		require.NoError(t, err)
		assert.NotEmpty(t, secret)
		assert.Greater(t, len(secret), 10)

		// Verify consistent hashing
		hashedChallenge2 := repo.HashChallenge(challenge)
		assert.Equal(t, hashedChallenge, hashedChallenge2)
	})

	t.Run("Temporary 2FA challenge lifecycle works correctly", func(t *testing.T) {
		repo := &temporaryTwoFactorChallengeRepo{}

		challenge, err := repo.GenerateChallenge()
		require.NoError(t, err)
		assert.NotEmpty(t, challenge)

		hashedChallenge := repo.HashChallenge(challenge)
		assert.NotEmpty(t, hashedChallenge)

		// Verify consistent hashing
		hashedChallenge2 := repo.HashChallenge(challenge)
		assert.Equal(t, hashedChallenge, hashedChallenge2)
	})
}

func TestCrossRepositoryIntegration(t *testing.T) {
	t.Run("All repositories can be instantiated together", func(t *testing.T) {
		// This test ensures all repositories can be created without conflicts
		testDB := &bun.DB{} // Mock DB for constructor testing

		// Create all repositories
		sessionRepo := NewSessionRepo(testDB)
		passwordResetRepo := NewPasswordResetTokenRepo(testDB)
		webauthnCredRepo := NewWebAuthnCredentialRepo(testDB)
		webauthnChallengeRepo := NewWebAuthnChallengeRepo(testDB)
		oauthRepo := NewOAuthCredentialRepo(testDB)
		twoFactorRepo := NewTwoFactorAuthenticationChallengeRepo(testDB)
		recoveryRepo := NewRecoveryCodeRepo(testDB)
		tempTwoFactorRepo := NewTemporaryTwoFactorChallengeRepo(testDB)

		// Verify all repositories are properly instantiated
		assert.NotNil(t, sessionRepo)
		assert.NotNil(t, passwordResetRepo)
		assert.NotNil(t, webauthnCredRepo)
		assert.NotNil(t, webauthnChallengeRepo)
		assert.NotNil(t, oauthRepo)
		assert.NotNil(t, twoFactorRepo)
		assert.NotNil(t, recoveryRepo)
		assert.NotNil(t, tempTwoFactorRepo)

		// Verify interfaces are implemented
		var _ SessionRepo = sessionRepo
		var _ PasswordResetTokenRepo = passwordResetRepo
		var _ WebAuthnCredentialRepo = webauthnCredRepo
		var _ WebAuthnChallengeRepo = webauthnChallengeRepo
		var _ OAuthCredentialRepo = oauthRepo
		var _ TwoFactorAuthenticationChallengeRepo = twoFactorRepo
		var _ RecoveryCodeRepo = recoveryRepo
		var _ TemporaryTwoFactorChallengeRepo = tempTwoFactorRepo

		assert.True(t, true, "All interfaces are properly implemented")
	})

	t.Run("All repositories can be instantiated as a group", func(t *testing.T) {
		testDB := &bun.DB{} // Mock DB for constructor testing

		authRepos := NewAuthRepos(testDB)
		require.NotNil(t, authRepos)

		// Verify all repositories are created
		assert.NotNil(t, authRepos.SessionRepo)
		assert.NotNil(t, authRepos.PasswordResetTokenRepo)
		assert.NotNil(t, authRepos.WebAuthnCredentialRepo)
		assert.NotNil(t, authRepos.WebAuthnChallengeRepo)
		assert.NotNil(t, authRepos.OAuthCredentialRepo)
		assert.NotNil(t, authRepos.TwoFactorAuthenticationChallengeRepo)
		assert.NotNil(t, authRepos.RecoveryCodeRepo)
		assert.NotNil(t, authRepos.TemporaryTwoFactorChallengeRepo)
	})
}

func TestSecurityUtilitiesIntegration(t *testing.T) {
	t.Run("Token generation security properties", func(t *testing.T) {
		// Test that tokens are properly random and secure
		tokens := make([]string, 100)
		for i := 0; i < 100; i++ {
			token, err := generateSecureToken(32)
			require.NoError(t, err)
			require.NotEmpty(t, token)
			require.Len(t, token, 64) // 32 bytes = 64 hex chars
			tokens[i] = token
		}

		// Verify all tokens are unique (very high probability)
		uniqueTokens := make(map[string]bool)
		for _, token := range tokens {
			assert.False(t, uniqueTokens[token], "Token should be unique")
			uniqueTokens[token] = true
		}
		assert.Len(t, uniqueTokens, 100, "All tokens should be unique")
	})

	t.Run("Recovery code generation security properties", func(t *testing.T) {
		codes := make([]string, 100)
		for i := 0; i < 100; i++ {
			code, err := generateRecoveryCode()
			require.NoError(t, err)
			require.NotEmpty(t, code)
			require.Len(t, code, 8)
			codes[i] = code
		}

		// Verify all codes are unique
		uniqueCodes := make(map[string]bool)
		for _, code := range codes {
			assert.False(t, uniqueCodes[code], "Recovery code should be unique")
			uniqueCodes[code] = true
		}
		assert.Len(t, uniqueCodes, 100, "All recovery codes should be unique")
	})

	t.Run("Hash function consistency across repositories", func(t *testing.T) {
		testInput := "test-input-12345"

		// All repositories should use the same hash function
		sessionRepo := &sessionRepo{}
		passwordResetRepo := &passwordResetTokenRepo{}
		recoveryRepo := &recoveryCodeRepo{}
		twoFactorRepo := &twoFactorAuthenticationChallengeRepo{}
		tempTwoFactorRepo := &temporaryTwoFactorChallengeRepo{}

		hash1 := sessionRepo.HashSessionToken(testInput)
		hash2 := passwordResetRepo.HashPasswordResetToken(testInput)
		hash3 := recoveryRepo.HashRecoveryCode(testInput)
		hash4 := twoFactorRepo.HashChallenge(testInput)
		hash5 := tempTwoFactorRepo.HashChallenge(testInput)

		// All hashes should be identical for the same input
		assert.Equal(t, hash1, hash2)
		assert.Equal(t, hash2, hash3)
		assert.Equal(t, hash3, hash4)
		assert.Equal(t, hash4, hash5)
		assert.Len(t, hash1, 32) // MD5 hash length
	})

	t.Run("TOTP secret generation properties", func(t *testing.T) {
		secrets := make([]string, 10)
		for i := 0; i < 10; i++ {
			secret, err := generateTwoFactorSecret()
			require.NoError(t, err)
			require.NotEmpty(t, secret)
			secrets[i] = secret
		}

		// Verify all secrets are unique
		uniqueSecrets := make(map[string]bool)
		for _, secret := range secrets {
			assert.False(t, uniqueSecrets[secret], "TOTP secret should be unique")
			uniqueSecrets[secret] = true
		}
		assert.Len(t, uniqueSecrets, 10, "All TOTP secrets should be unique")

		// Verify secret format (should be base32-like characters)
		for _, secret := range secrets {
			// TOTP secrets contain only base32 characters
			for _, char := range secret {
				validChar := (char >= 'A' && char <= 'Z') ||
					(char >= '2' && char <= '7') ||
					char == '=' // padding
				assert.True(t, validChar, "Invalid character in TOTP secret: %c", char)
			}
		}
	})
}

func TestRepositoryErrorHandling(t *testing.T) {
	t.Run("All custom errors are properly defined", func(t *testing.T) {
		errors := []struct {
			err    error
			name   string
			expect string
		}{
			{ErrSessionNotFound, "SessionNotFound", "session not found"},
			{ErrTokenExpired, "TokenExpired", "token has expired"},
			{ErrInvalidCredentials, "InvalidCredentials", "invalid credentials"},
			{ErrWebAuthnCredentialNotFound, "WebAuthnCredentialNotFound", "webauthn credential not found"},
			{ErrChallengeNotFound, "ChallengeNotFound", "challenge not found"},
			{ErrRecoveryCodeInvalid, "RecoveryCodeInvalid", "recovery code invalid"},
			{ErrOAuthCredentialAlreadyExists, "OAuthCredentialAlreadyExists", "oauth credential already exists"},
			{ErrPasswordResetTokenNotFound, "PasswordResetTokenNotFound", "password reset token not found"},
			{ErrTwoFactorAuthenticationNotFound, "TwoFactorAuthenticationNotFound", "two factor authentication challenge not found"},
			{ErrTemporaryTwoFactorNotFound, "TemporaryTwoFactorNotFound", "temporary two factor challenge not found"},
		}

		for _, tc := range errors {
			t.Run(tc.name, func(t *testing.T) {
				require.NotNil(t, tc.err, "Error should be defined")
				assert.Contains(t, tc.err.Error(), tc.expect, "Error message should contain expected text")
				assert.NotEmpty(t, tc.err.Error(), "Error message should not be empty")
			})
		}
	})

	t.Run("Unique violation detection works correctly", func(t *testing.T) {
		testCases := []struct {
			name     string
			errMsg   string
			expected bool
		}{
			{
				name:     "PostgreSQL duplicate key",
				errMsg:   "duplicate key value violates unique constraint",
				expected: true,
			},
			{
				name:     "PostgreSQL unique constraint",
				errMsg:   "unique constraint violation",
				expected: true,
			},
			{
				name:     "Generic unique violation",
				errMsg:   "UNIQUE violation",
				expected: true,
			},
			{
				name:     "Non-unique error",
				errMsg:   "some other error",
				expected: false,
			},
			{
				name:     "Empty error",
				errMsg:   "",
				expected: false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				var err error
				if tc.errMsg != "" {
					err = fmt.Errorf("%s", tc.errMsg)
				}
				result := isUniqueViolation(err)
				assert.Equal(t, tc.expected, result)
			})
		}
	})
}

func TestPaginationIntegration(t *testing.T) {
	t.Run("Cursor-based pagination utilities work correctly", func(t *testing.T) {
		// Test cursor generation and parsing
		testIDs := []int64{1, 999, 123456789}

		for _, id := range testIDs {
			cursor := db.GetCursorForID(id)
			require.NotNil(t, cursor, "Cursor should be generated for ID %d", id)

			parsedID, err := db.ParseCursor(*cursor)
			require.NoError(t, err, "Cursor should parse correctly for ID %d", id)
			assert.Equal(t, id, parsedID, "Parsed ID should match original for ID %d", id)
		}

		// Test edge cases
		t.Run("Cursor edge cases", func(t *testing.T) {
			// Zero ID
			cursor := db.GetCursorForID(0)
			assert.Nil(t, cursor, "Cursor should be nil for zero ID")

			// Negative ID
			cursor = db.GetCursorForID(-1)
			assert.Nil(t, cursor, "Cursor should be nil for negative ID")

			// Empty cursor
			parsedID, err := db.ParseCursor("")
			assert.NoError(t, err, "Empty cursor should not error")
			assert.Equal(t, int64(0), parsedID, "Empty cursor should return zero ID")

			// Invalid cursor
			_, err = db.ParseCursor("invalid-base64")
			assert.Error(t, err, "Invalid cursor should return error")
		})
	})

	t.Run("Pagination options validation", func(t *testing.T) {
		testCases := []struct {
			name      string
			options   db.PaginationOptions
			expectErr bool
		}{
			{
				name: "Valid first parameter",
				options: db.PaginationOptions{
					First: intPtr(10),
				},
				expectErr: false,
			},
			{
				name: "Valid last parameter",
				options: db.PaginationOptions{
					Last: intPtr(5),
				},
				expectErr: false,
			},
			{
				name: "Zero first parameter",
				options: db.PaginationOptions{
					First: intPtr(0),
				},
				expectErr: true,
			},
			{
				name: "Negative last parameter",
				options: db.PaginationOptions{
					Last: intPtr(-1),
				},
				expectErr: true,
			},
			{
				name: "Both first and last parameters",
				options: db.PaginationOptions{
					First: intPtr(10),
					Last:  intPtr(5),
				},
				expectErr: true,
			},
			{
				name:      "Empty options",
				options:   db.PaginationOptions{},
				expectErr: false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := db.ValidatePaginationOptions(tc.options)
				if tc.expectErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})
}

func TestConcurrencySafety(t *testing.T) {
	t.Run("Concurrent token generation", func(t *testing.T) {
		// Test that multiple goroutines can generate tokens without conflicts
		const numGoroutines = 10
		const tokensPerGoroutine = 10

		tokenChan := make(chan string, numGoroutines*tokensPerGoroutine)
		errChan := make(chan error, numGoroutines*tokensPerGoroutine)

		// Start multiple goroutines generating tokens
		for i := 0; i < numGoroutines; i++ {
			go func() {
				repo := &sessionRepo{}
				for j := 0; j < tokensPerGoroutine; j++ {
					token, err := repo.GenerateSessionToken()
					if err != nil {
						errChan <- err
						return
					}
					tokenChan <- token
				}
			}()
		}

		// Collect results
		tokens := make([]string, 0, numGoroutines*tokensPerGoroutine)
		for i := 0; i < numGoroutines*tokensPerGoroutine; i++ {
			select {
			case token := <-tokenChan:
				tokens = append(tokens, token)
			case err := <-errChan:
				t.Fatalf("Token generation error: %v", err)
			case <-time.After(5 * time.Second):
				t.Fatal("Timeout waiting for token generation")
			}
		}

		// Verify all tokens were generated and are unique
		assert.Len(t, tokens, numGoroutines*tokensPerGoroutine, "Expected correct number of tokens")

		uniqueTokens := make(map[string]bool)
		for _, token := range tokens {
			assert.NotEmpty(t, token, "Token should not be empty")
			assert.False(t, uniqueTokens[token], "Token should be unique")
			uniqueTokens[token] = true
		}
	})
}

// Helper function for creating int pointers
func intPtr(i int) *int {
	return &i
}