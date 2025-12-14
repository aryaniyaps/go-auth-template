package account

import (
	"context"
	"errors"
	"testing"

	"server/internal/domain/core"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestAccountService_UpdateAccountTermsAndPolicy(t *testing.T) {
	logger := zap.NewNop()
	databaseError := errors.New("database connection error")

	tests := []struct {
		name         string
		accountID    int64
		version      string
		setupMocks   func(*MockAccountRepo)
		expectError  bool
		errorMessage string
	}{
		{
			name:      "successful terms update with version tracking",
			accountID: 1,
			version:   "2.1.0",
			setupMocks: func(m *MockAccountRepo) {
				testAccount := &Account{
					CoreModel: core.CoreModel{ID: 1},
					FullName:  "John Doe",
					Email:     "john@example.com",
				}
				updatedAccount := &Account{
					CoreModel: core.CoreModel{ID: 1},
					FullName:  "John Doe",
					Email:     "john@example.com",
					TermsAndPolicy: TermsAndPolicy{
						Type:    "accepted",
						Version: "2.1.0",
					},
				}
				m.On("Get", mock.Anything, int64(1)).Return(testAccount, nil)
				m.On("Update", mock.Anything, testAccount, (*string)(nil), (*string)(nil), (*string)(nil), mock.AnythingOfType("*account.TermsAndPolicy"), (*AnalyticsPreference)(nil), (*bool)(nil)).Return(updatedAccount, nil)
			},
			expectError: false,
		},
		{
			name:         "empty version should return error",
			accountID:    1,
			version:      "",
			setupMocks:   func(m *MockAccountRepo) {},
			expectError:  true,
			errorMessage: "invalid input: terms version cannot be empty",
		},
		{
			name:         "whitespace-only version should return error",
			accountID:    1,
			version:      "   ",
			setupMocks:   func(m *MockAccountRepo) {},
			expectError:  true,
			errorMessage: "invalid input: terms version cannot be empty",
		},
		{
			name:      "account not found should return error",
			accountID: 999,
			version:   "1.0.0",
			setupMocks: func(m *MockAccountRepo) {
				m.On("Get", mock.Anything, int64(999)).Return(nil, ErrAccountNotFound)
			},
			expectError:  true,
			errorMessage: "failed to get account: account not found",
		},
		{
			name:      "database error during update should return error",
			accountID: 1,
			version:   "3.0.0",
			setupMocks: func(m *MockAccountRepo) {
				testAccount := &Account{
					CoreModel: core.CoreModel{ID: 1},
					FullName:  "John Doe",
					Email:     "john@example.com",
				}
				m.On("Get", mock.Anything, int64(1)).Return(testAccount, nil)
				m.On("Update", mock.Anything, testAccount, (*string)(nil), (*string)(nil), (*string)(nil), mock.AnythingOfType("*account.TermsAndPolicy"), (*AnalyticsPreference)(nil), (*bool)(nil)).Return(nil, databaseError)
			},
			expectError:  true,
			errorMessage: "failed to update account terms and policy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh mocks for each test to avoid state conflicts
			mockRepo := &MockAccountRepo{}
			mockSMS := &MockMessageSender{}
			service := NewAccountService(mockRepo, nil, nil, mockSMS, nil, logger)

			tt.setupMocks(mockRepo)

			ctx := context.Background()
			result, err := service.UpdateAccountTermsAndPolicy(ctx, tt.accountID, tt.version)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.version, result.TermsAndPolicy.Version)
				assert.Equal(t, "accepted", result.TermsAndPolicy.Type)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAccountService_UpdateAccountAnalyticsPreference(t *testing.T) {
	logger := zap.NewNop()
	databaseError := errors.New("database connection error")

	tests := []struct {
		name         string
		accountID    int64
		preference   string
		setupMocks   func(*MockAccountRepo)
		expectError  bool
		errorMessage string
	}{
		{
			name:       "successful analytics preference update - enabled",
			accountID:  1,
			preference: "enabled",
			setupMocks: func(m *MockAccountRepo) {
				testAccount := &Account{
					CoreModel: core.CoreModel{ID: 1},
					FullName:  "John Doe",
					Email:     "john@example.com",
				}
				updatedAccount := &Account{
					CoreModel: core.CoreModel{ID: 1},
					FullName:  "John Doe",
					Email:     "john@example.com",
					AnalyticsPref: AnalyticsPreference{
						Type: "enabled",
					},
				}
				m.On("Get", mock.Anything, int64(1)).Return(testAccount, nil)
				m.On("Update", mock.Anything, testAccount, (*string)(nil), (*string)(nil), (*string)(nil), (*TermsAndPolicy)(nil), mock.AnythingOfType("*account.AnalyticsPreference"), (*bool)(nil)).Return(updatedAccount, nil)
			},
			expectError: false,
		},
		{
			name:       "successful analytics preference update - disabled",
			accountID:  1,
			preference: "disabled",
			setupMocks: func(m *MockAccountRepo) {
				testAccount := &Account{
					CoreModel: core.CoreModel{ID: 1},
					FullName:  "John Doe",
					Email:     "john@example.com",
				}
				updatedAccount := &Account{
					CoreModel: core.CoreModel{ID: 1},
					FullName:  "John Doe",
					Email:     "john@example.com",
					AnalyticsPref: AnalyticsPreference{
						Type: "disabled",
					},
				}
				m.On("Get", mock.Anything, int64(1)).Return(testAccount, nil)
				m.On("Update", mock.Anything, testAccount, (*string)(nil), (*string)(nil), (*string)(nil), (*TermsAndPolicy)(nil), mock.AnythingOfType("*account.AnalyticsPreference"), (*bool)(nil)).Return(updatedAccount, nil)
			},
			expectError: false,
		},
		{
			name:       "successful analytics preference update - undecided",
			accountID:  1,
			preference: "undecided",
			setupMocks: func(m *MockAccountRepo) {
				testAccount := &Account{
					CoreModel: core.CoreModel{ID: 1},
					FullName:  "John Doe",
					Email:     "john@example.com",
				}
				updatedAccount := &Account{
					CoreModel: core.CoreModel{ID: 1},
					FullName:  "John Doe",
					Email:     "john@example.com",
					AnalyticsPref: AnalyticsPreference{
						Type: "undecided",
					},
				}
				m.On("Get", mock.Anything, int64(1)).Return(testAccount, nil)
				m.On("Update", mock.Anything, testAccount, (*string)(nil), (*string)(nil), (*string)(nil), (*TermsAndPolicy)(nil), mock.AnythingOfType("*account.AnalyticsPreference"), (*bool)(nil)).Return(updatedAccount, nil)
			},
			expectError: false,
		},
		{
			name:         "invalid preference value should return error",
			accountID:    1,
			preference:   "invalid",
			setupMocks:   func(m *MockAccountRepo) {},
			expectError:  true,
			errorMessage: "invalid preference value: preference must be 'enabled', 'disabled', or 'undecided'",
		},
		{
			name:       "account not found should return error",
			accountID:  999,
			preference: "enabled",
			setupMocks: func(m *MockAccountRepo) {
				m.On("Get", mock.Anything, int64(999)).Return(nil, ErrAccountNotFound)
			},
			expectError:  true,
			errorMessage: "failed to get account: account not found",
		},
		{
			name:       "database error during update should return error",
			accountID:  1,
			preference: "disabled",
			setupMocks: func(m *MockAccountRepo) {
				testAccount := &Account{
					CoreModel: core.CoreModel{ID: 1},
					FullName:  "John Doe",
					Email:     "john@example.com",
				}
				m.On("Get", mock.Anything, int64(1)).Return(testAccount, nil)
				m.On("Update", mock.Anything, testAccount, (*string)(nil), (*string)(nil), (*string)(nil), (*TermsAndPolicy)(nil), mock.AnythingOfType("*account.AnalyticsPreference"), (*bool)(nil)).Return(nil, databaseError)
			},
			expectError:  true,
			errorMessage: "failed to update account analytics preference",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh mocks for each test to avoid state conflicts
			mockRepo := &MockAccountRepo{}
			mockSMS := &MockMessageSender{}
			service := NewAccountService(mockRepo, nil, nil, mockSMS, nil, logger)

			tt.setupMocks(mockRepo)

			ctx := context.Background()
			result, err := service.UpdateAccountAnalyticsPreference(ctx, tt.accountID, tt.preference)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.preference, result.AnalyticsPref.Type)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAccountService_UpdateAccountWhatsappJobAlerts(t *testing.T) {
	logger := zap.NewNop()
	databaseError := errors.New("database connection error")

	tests := []struct {
		name         string
		accountID    int64
		enabled      bool
		setupMocks   func(*MockAccountRepo)
		expectError  bool
		errorMessage string
	}{
		{
			name:      "successful WhatsApp job alerts update - enable",
			accountID: 1,
			enabled:   true,
			setupMocks: func(m *MockAccountRepo) {
				testAccount := &Account{
					CoreModel: core.CoreModel{ID: 1},
					FullName:  "John Doe",
					Email:     "john@example.com",
				}
				updatedAccount := &Account{
					CoreModel: core.CoreModel{ID: 1},
					FullName:  "John Doe",
					Email:     "john@example.com",
				}
				m.On("Get", mock.Anything, int64(1)).Return(testAccount, nil)
				m.On("Update", mock.Anything, testAccount, (*string)(nil), (*string)(nil), (*string)(nil), (*TermsAndPolicy)(nil), (*AnalyticsPreference)(nil), mock.AnythingOfType("*bool")).Return(updatedAccount, nil)
			},
			expectError: false,
		},
		{
			name:      "successful WhatsApp job alerts update - disable",
			accountID: 1,
			enabled:   false,
			setupMocks: func(m *MockAccountRepo) {
				testAccount := &Account{
					CoreModel: core.CoreModel{ID: 1},
					FullName:  "John Doe",
					Email:     "john@example.com",
				}
				updatedAccount := &Account{
					CoreModel: core.CoreModel{ID: 1},
					FullName:  "John Doe",
					Email:     "john@example.com",
				}
				m.On("Get", mock.Anything, int64(1)).Return(testAccount, nil)
				m.On("Update", mock.Anything, testAccount, (*string)(nil), (*string)(nil), (*string)(nil), (*TermsAndPolicy)(nil), (*AnalyticsPreference)(nil), mock.AnythingOfType("*bool")).Return(updatedAccount, nil)
			},
			expectError: false,
		},
		{
			name:      "account not found should return error",
			accountID: 999,
			enabled:   true,
			setupMocks: func(m *MockAccountRepo) {
				m.On("Get", mock.Anything, int64(999)).Return(nil, ErrAccountNotFound)
			},
			expectError:  true,
			errorMessage: "failed to get account: account not found",
		},
		{
			name:      "database error during update should return error",
			accountID: 1,
			enabled:   true,
			setupMocks: func(m *MockAccountRepo) {
				testAccount := &Account{
					CoreModel: core.CoreModel{ID: 1},
					FullName:  "John Doe",
					Email:     "john@example.com",
				}
				m.On("Get", mock.Anything, int64(1)).Return(testAccount, nil)
				m.On("Update", mock.Anything, testAccount, (*string)(nil), (*string)(nil), (*string)(nil), (*TermsAndPolicy)(nil), (*AnalyticsPreference)(nil), mock.AnythingOfType("*bool")).Return(nil, databaseError)
			},
			expectError:  true,
			errorMessage: "failed to update account WhatsApp job alerts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh mocks for each test to avoid state conflicts
			mockRepo := &MockAccountRepo{}
			mockSMS := &MockMessageSender{}
			service := NewAccountService(mockRepo, nil, nil, mockSMS, nil, logger)

			tt.setupMocks(mockRepo)

			ctx := context.Background()
			result, err := service.UpdateAccountWhatsappJobAlerts(ctx, tt.accountID, tt.enabled)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAccountService_AdvancedMethodsIntegration(t *testing.T) {
	mockRepo := &MockAccountRepo{}
	mockSMS := &MockMessageSender{}
	logger := zap.NewNop()
	service := NewAccountService(mockRepo, nil, nil, mockSMS, nil, logger)

	testAccount := &Account{
		CoreModel: core.CoreModel{ID: 1},
		FullName:  "John Doe",
		Email:     "john@example.com",
	}

	// Test: Update terms and policy with structured data capture
	var capturedTermsAndPolicy *TermsAndPolicy
	mockRepo.On("Get", mock.Anything, int64(1)).Return(testAccount, nil)
	mockRepo.On("Update", mock.Anything, testAccount, (*string)(nil), (*string)(nil), (*string)(nil), mock.AnythingOfType("*account.TermsAndPolicy"), (*AnalyticsPreference)(nil), (*bool)(nil)).Return(testAccount, nil).Run(func(args mock.Arguments) {
		capturedTermsAndPolicy = args.Get(5).(*TermsAndPolicy)
	})

	ctx := context.Background()
	result, err := service.UpdateAccountTermsAndPolicy(ctx, 1, "2.1.0")

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, capturedTermsAndPolicy)
	assert.Equal(t, "accepted", capturedTermsAndPolicy.Type)
	assert.Equal(t, "2.1.0", capturedTermsAndPolicy.Version)
	assert.NotZero(t, capturedTermsAndPolicy.UpdatedAt)

	mockRepo.AssertExpectations(t)

	// Test: Update analytics preference with structured data capture
	mockRepo = &MockAccountRepo{} // Create fresh mock
	service = NewAccountService(mockRepo, nil, nil, mockSMS, nil, logger)

	var capturedAnalyticsPreference *AnalyticsPreference
	mockRepo.On("Get", mock.Anything, int64(1)).Return(testAccount, nil)
	mockRepo.On("Update", mock.Anything, testAccount, (*string)(nil), (*string)(nil), (*string)(nil), (*TermsAndPolicy)(nil), mock.AnythingOfType("*account.AnalyticsPreference"), (*bool)(nil)).Return(testAccount, nil).Run(func(args mock.Arguments) {
		capturedAnalyticsPreference = args.Get(6).(*AnalyticsPreference)
	})

	result, err = service.UpdateAccountAnalyticsPreference(ctx, 1, "enabled")

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, capturedAnalyticsPreference)
	assert.Equal(t, "enabled", capturedAnalyticsPreference.Type)
	assert.NotZero(t, capturedAnalyticsPreference.UpdatedAt)

	mockRepo.AssertExpectations(t)
}
