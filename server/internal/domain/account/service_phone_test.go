package account

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"server/internal/domain/core"
	"go.uber.org/zap"
)

func TestAccountService_CreatePhoneVerificationToken(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name         string
		phoneNumber  string
		setupMocks   func(*MockAccountRepo, *MockPhoneNumberVerificationTokenRepo, *MockMessageSender)
		expectError  bool
		errorMessage string
	}{
		{
			name:        "successful phone verification token creation",
			phoneNumber: "+12345678901",
			setupMocks: func(m *MockAccountRepo, p *MockPhoneNumberVerificationTokenRepo, s *MockMessageSender) {
				testPhoneToken := &PhoneNumberVerificationToken{
					CoreModel:  core.CoreModel{ID: 1},
					PhoneNumber: "+12345678901",
					TokenHash:   "hashedtoken123",
					ExpiresAt:   time.Now().Add(24 * time.Hour),
				}
				p.On("Create", mock.Anything, "+12345678901").Return("123456", testPhoneToken, nil)
				s.On("SendSMS", mock.Anything, "+12345678901", mock.AnythingOfType("string")).Return(nil)
			},
			expectError: false,
		},
		{
			name:         "invalid phone number should return error",
			phoneNumber: "invalid",
			setupMocks:   func(m *MockAccountRepo, p *MockPhoneNumberVerificationTokenRepo, s *MockMessageSender) {},
			expectError:  true,
			errorMessage: "invalid phone number format",
		},
		{
			name:         "empty phone number should return error",
			phoneNumber: "",
			setupMocks:   func(m *MockAccountRepo, p *MockPhoneNumberVerificationTokenRepo, s *MockMessageSender) {},
			expectError:  true,
			errorMessage: "invalid phone number format",
		},
		{
			name:        "failed to create token should return error",
			phoneNumber: "+12345678901",
			setupMocks: func(m *MockAccountRepo, p *MockPhoneNumberVerificationTokenRepo, s *MockMessageSender) {
				p.On("Create", mock.Anything, "+12345678901").Return("", (*PhoneNumberVerificationToken)(nil), assert.AnError)
			},
			expectError:  true,
			errorMessage: "failed to create verification token",
		},
		{
			name:        "failed to send SMS should return error",
			phoneNumber: "+12345678901",
			setupMocks: func(m *MockAccountRepo, p *MockPhoneNumberVerificationTokenRepo, s *MockMessageSender) {
				testPhoneToken := &PhoneNumberVerificationToken{
					CoreModel:  core.CoreModel{ID: 1},
					PhoneNumber: "+12345678901",
					TokenHash:   "hashedtoken123",
					ExpiresAt:   time.Now().Add(24 * time.Hour),
				}
				p.On("Create", mock.Anything, "+12345678901").Return("123456", testPhoneToken, nil)
				s.On("SendSMS", mock.Anything, "+12345678901", mock.AnythingOfType("string")).Return(assert.AnError)
			},
			expectError:  true,
			errorMessage: "failed to send SMS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh mocks for each test to avoid state conflicts
			mockRepo := &MockAccountRepo{}
			mockPhoneTokenRepo := &MockPhoneNumberVerificationTokenRepo{}
			mockMessageSender := &MockMessageSender{}
			service := NewAccountService(mockRepo, mockPhoneTokenRepo, nil, mockMessageSender, nil, logger)

			tt.setupMocks(mockRepo, mockPhoneTokenRepo, mockMessageSender)

			ctx := context.Background()
			err := service.CreatePhoneVerificationToken(ctx, tt.phoneNumber)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
			} else {
				require.NoError(t, err)
			}

			mockPhoneTokenRepo.AssertExpectations(t)
			mockMessageSender.AssertExpectations(t)
		})
	}
}

func TestAccountService_VerifyPhoneNumber(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name         string
		phoneNumber  string
		token        string
		setupMocks   func(*MockAccountRepo, *MockPhoneNumberVerificationTokenRepo)
		expectError  bool
		errorMessage string
	}{
		{
			name:        "successful phone number verification with existing account",
			phoneNumber: "+12345678901",
			token:       "123456",
			setupMocks: func(m *MockAccountRepo, p *MockPhoneNumberVerificationTokenRepo) {
				// Mock GetByPhoneNumber for token
				testPhoneToken := &PhoneNumberVerificationToken{
					CoreModel:  core.CoreModel{ID: 1},
					PhoneNumber: "+12345678901",
					TokenHash:   "e10adc3949ba59abbe56e057f20f883e", // MD5 of "123456"
					ExpiresAt:   time.Now().Add(1 * time.Hour), // Not expired
				}
				p.On("GetByPhoneNumber", mock.Anything, "+12345678901").Return(testPhoneToken, nil)
				p.On("HashVerificationToken", "123456").Return("e10adc3949ba59abbe56e057f20f883e")

				// Mock GetByPhoneNumber for account
				testAccount := &Account{
					CoreModel: core.CoreModel{ID: 1},
					FullName:  "John Doe",
					Email:     "john@example.com",
				}
				m.On("GetByPhoneNumber", mock.Anything, "+12345678901").Return(testAccount, nil)

				// Mock Update
				updatedAccount := &Account{
					CoreModel:   core.CoreModel{ID: 1},
					FullName:    "John Doe",
					Email:       "john@example.com",
					PhoneNumber: stringPtr("+12345678901"),
				}
				m.On("Update", mock.Anything, testAccount, (*string)(nil), (*string)(nil), mock.AnythingOfType("*string"), (*TermsAndPolicy)(nil), (*AnalyticsPreference)(nil), (*bool)(nil)).Return(updatedAccount, nil)

				// Mock Delete
				p.On("Delete", mock.Anything, testPhoneToken).Return(nil)
			},
			expectError: false,
		},
		{
			name:        "successful phone number verification without existing account",
			phoneNumber: "+12345678901",
			token:       "123456",
			setupMocks: func(m *MockAccountRepo, p *MockPhoneNumberVerificationTokenRepo) {
				// Mock GetByPhoneNumber for token
				testPhoneToken := &PhoneNumberVerificationToken{
					CoreModel:  core.CoreModel{ID: 1},
					PhoneNumber: "+12345678901",
					TokenHash:   "e10adc3949ba59abbe56e057f20f883e", // MD5 of "123456"
					ExpiresAt:   time.Now().Add(1 * time.Hour), // Not expired
				}
				p.On("GetByPhoneNumber", mock.Anything, "+12345678901").Return(testPhoneToken, nil)
				p.On("HashVerificationToken", "123456").Return("e10adc3949ba59abbe56e057f20f883e")

				// Mock GetByPhoneNumber for account (not found)
				m.On("GetByPhoneNumber", mock.Anything, "+12345678901").Return((*Account)(nil), ErrAccountNotFound)

				// Mock Delete
				p.On("Delete", mock.Anything, testPhoneToken).Return(nil)
			},
			expectError: false,
		},
		{
			name:         "empty phone number should return error",
			phoneNumber:  "",
			token:        "123456",
			setupMocks:   func(m *MockAccountRepo, p *MockPhoneNumberVerificationTokenRepo) {},
			expectError:  true,
			errorMessage: "invalid input: phone number cannot be empty",
		},
		{
			name:         "empty token should return error",
			phoneNumber:  "+12345678901",
			token:        "",
			setupMocks:   func(m *MockAccountRepo, p *MockPhoneNumberVerificationTokenRepo) {},
			expectError:  true,
			errorMessage: "invalid input: token cannot be empty",
		},
		{
			name:         "invalid phone number should return error",
			phoneNumber:  "invalid",
			token:        "123456",
			setupMocks:   func(m *MockAccountRepo, p *MockPhoneNumberVerificationTokenRepo) {},
			expectError:  true,
			errorMessage: "invalid phone number format",
		},
		{
			name:        "token not found should return error",
			phoneNumber: "+12345678901",
			token:       "123456",
			setupMocks: func(m *MockAccountRepo, p *MockPhoneNumberVerificationTokenRepo) {
				p.On("GetByPhoneNumber", mock.Anything, "+12345678901").Return((*PhoneNumberVerificationToken)(nil), ErrTokenNotFound)
			},
			expectError:  true,
			errorMessage: "invalid verification token: no verification token found for this phone number",
		},
		{
			name:        "expired token should return error",
			phoneNumber: "+12345678901",
			token:       "123456",
			setupMocks: func(m *MockAccountRepo, p *MockPhoneNumberVerificationTokenRepo) {
				// Create expired token (expired 1 hour ago)
				testPhoneToken := &PhoneNumberVerificationToken{
					CoreModel:  core.CoreModel{ID: 1},
					PhoneNumber: "+12345678901",
					TokenHash:   "e10adc3949ba59abbe56e057f20f883e", // MD5 of "123456"
					ExpiresAt:   time.Now().Add(-1 * time.Hour), // Expired
				}
				p.On("GetByPhoneNumber", mock.Anything, "+12345678901").Return(testPhoneToken, nil)
			},
			expectError:  true,
			errorMessage: "token has expired",
		},
		{
			name:        "invalid token should return error",
			phoneNumber: "+12345678901",
			token:       "wrongtoken",
			setupMocks: func(m *MockAccountRepo, p *MockPhoneNumberVerificationTokenRepo) {
				// Create token with hash for "123456"
				testPhoneToken := &PhoneNumberVerificationToken{
					CoreModel:  core.CoreModel{ID: 1},
					PhoneNumber: "+12345678901",
					TokenHash:   "e10adc3949ba59abbe56e057f20f883e", // MD5 of "123456"
					ExpiresAt:   time.Now().Add(1 * time.Hour), // Not expired
				}
				p.On("GetByPhoneNumber", mock.Anything, "+12345678901").Return(testPhoneToken, nil)
				p.On("HashVerificationToken", "wrongtoken").Return("7d865e959b2466918c9863afca942d0fb8d4c8a8")
			},
			expectError:  true,
			errorMessage: "invalid verification token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh mocks for each test to avoid state conflicts
			mockRepo := &MockAccountRepo{}
			mockPhoneTokenRepo := &MockPhoneNumberVerificationTokenRepo{}
			service := NewAccountService(mockRepo, mockPhoneTokenRepo, nil, nil, nil, logger)

			tt.setupMocks(mockRepo, mockPhoneTokenRepo)

			ctx := context.Background()
			err := service.VerifyPhoneNumber(ctx, tt.phoneNumber, tt.token)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
			} else {
				require.NoError(t, err)
			}

			mockPhoneTokenRepo.AssertExpectations(t)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAccountService_PhoneVerificationIntegration(t *testing.T) {
	logger := zap.NewNop()
	mockRepo := &MockAccountRepo{}
	mockPhoneTokenRepo := &MockPhoneNumberVerificationTokenRepo{}
	mockMessageSender := &MockMessageSender{}
	service := NewAccountService(mockRepo, mockPhoneTokenRepo, nil, mockMessageSender, nil, logger)

	// Test complete phone verification workflow
	ctx := context.Background()
	phoneNumber := "+12345678901"
	token := "123456"

	// Step 1: Setup mocks for token creation
	testPhoneToken := &PhoneNumberVerificationToken{
		CoreModel:  core.CoreModel{ID: 1},
		PhoneNumber: phoneNumber,
		TokenHash:   "e10adc3949ba59abbe56e057f20f883e", // MD5 of "123456"
		ExpiresAt:   time.Now().Add(24 * time.Hour),
	}
	mockPhoneTokenRepo.On("Create", mock.Anything, phoneNumber).Return(token, testPhoneToken, nil)
	mockMessageSender.On("SendSMS", mock.Anything, phoneNumber, mock.AnythingOfType("string")).Return(nil)

	// Step 2: Create verification token
	err := service.CreatePhoneVerificationToken(ctx, phoneNumber)
	require.NoError(t, err)

	// Step 3: Setup mocks for token verification (without existing account)
	mockPhoneTokenRepo.On("GetByPhoneNumber", mock.Anything, phoneNumber).Return(testPhoneToken, nil)
	mockPhoneTokenRepo.On("HashVerificationToken", token).Return("e10adc3949ba59abbe56e057f20f883e")
	mockRepo.On("GetByPhoneNumber", mock.Anything, phoneNumber).Return((*Account)(nil), ErrAccountNotFound)
	mockPhoneTokenRepo.On("Delete", mock.Anything, testPhoneToken).Return(nil)

	// Step 4: Verify phone number
	err = service.VerifyPhoneNumber(ctx, phoneNumber, token)
	require.NoError(t, err)

	// Verify all mocks were called
	mockPhoneTokenRepo.AssertExpectations(t)
	mockMessageSender.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestAccountService_PhoneVerificationTokenDeletionError(t *testing.T) {
	logger := zap.NewNop()
	mockRepo := &MockAccountRepo{}
	mockPhoneTokenRepo := &MockPhoneNumberVerificationTokenRepo{}
	service := NewAccountService(mockRepo, mockPhoneTokenRepo, nil, nil, nil, logger)

	// Test that deletion errors are logged but don't fail the verification
	ctx := context.Background()
	phoneNumber := "+12345678901"
	token := "123456"

	// Setup mocks
	testPhoneToken := &PhoneNumberVerificationToken{
		CoreModel:  core.CoreModel{ID: 1},
		PhoneNumber: phoneNumber,
		TokenHash:   "e10adc3949ba59abbe56e057f20f883e",
		ExpiresAt:   time.Now().Add(1 * time.Hour),
	}
	mockPhoneTokenRepo.On("GetByPhoneNumber", mock.Anything, phoneNumber).Return(testPhoneToken, nil)
	mockPhoneTokenRepo.On("HashVerificationToken", token).Return("e10adc3949ba59abbe56e057f20f883e")
	mockRepo.On("GetByPhoneNumber", mock.Anything, phoneNumber).Return((*Account)(nil), ErrAccountNotFound)
	mockPhoneTokenRepo.On("Delete", mock.Anything, testPhoneToken).Return(assert.AnError)

	// Verify phone number - should succeed despite deletion error
	err := service.VerifyPhoneNumber(ctx, phoneNumber, token)
	require.NoError(t, err)

	mockPhoneTokenRepo.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}