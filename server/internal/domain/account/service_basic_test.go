package account

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"server/internal/domain/core"
)

// MockAccountRepo is a mock implementation of AccountRepo for testing
type MockAccountRepo struct {
	mock.Mock
}

func (m *MockAccountRepo) Create(ctx context.Context, email string, fullName string, authProviders []string, password *string, accountID *int64, analyticsPreference string, phoneNumber *string) (*Account, error) {
	args := m.Called(ctx, email, fullName, authProviders, password, accountID, analyticsPreference, phoneNumber)
	return args.Get(0).(*Account), args.Error(1)
}

func (m *MockAccountRepo) Get(ctx context.Context, accountID int64) (*Account, error) {
	args := m.Called(ctx, accountID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Account), args.Error(1)
}

func (m *MockAccountRepo) GetByEmail(ctx context.Context, email string) (*Account, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Account), args.Error(1)
}

func (m *MockAccountRepo) GetByPhoneNumber(ctx context.Context, phoneNumber string) (*Account, error) {
	args := m.Called(ctx, phoneNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Account), args.Error(1)
}

func (m *MockAccountRepo) Update(ctx context.Context, account *Account, fullName *string, avatarURL *string, phoneNumber *string, termsAndPolicy *TermsAndPolicy, analyticsPreference *AnalyticsPreference, whatsappJobAlerts *bool) (*Account, error) {
	args := m.Called(ctx, account, fullName, avatarURL, phoneNumber, termsAndPolicy, analyticsPreference, whatsappJobAlerts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Account), args.Error(1)
}

func (m *MockAccountRepo) UpdateProfile(ctx context.Context, account *Account, profile any) (*Account, error) {
	args := m.Called(ctx, account, profile)
	return args.Get(0).(*Account), args.Error(1)
}

func (m *MockAccountRepo) UpdateAuthProviders(ctx context.Context, account *Account, authProviders []string) (*Account, error) {
	args := m.Called(ctx, account, authProviders)
	return args.Get(0).(*Account), args.Error(1)
}

func (m *MockAccountRepo) DeleteAvatar(ctx context.Context, account *Account) (*Account, error) {
	args := m.Called(ctx, account)
	return args.Get(0).(*Account), args.Error(1)
}

func (m *MockAccountRepo) SetTwoFactorSecret(ctx context.Context, account *Account, totpSecret string) (*Account, error) {
	args := m.Called(ctx, account, totpSecret)
	return args.Get(0).(*Account), args.Error(1)
}

func (m *MockAccountRepo) DeleteTwoFactorSecret(ctx context.Context, account *Account) (*Account, error) {
	args := m.Called(ctx, account)
	return args.Get(0).(*Account), args.Error(1)
}

func (m *MockAccountRepo) UpdatePassword(ctx context.Context, account *Account, password string) (*Account, error) {
	args := m.Called(ctx, account, password)
	return args.Get(0).(*Account), args.Error(1)
}

func (m *MockAccountRepo) DeletePassword(ctx context.Context, account *Account) (*Account, error) {
	args := m.Called(ctx, account)
	return args.Get(0).(*Account), args.Error(1)
}

func (m *MockAccountRepo) Delete(ctx context.Context, account *Account) error {
	args := m.Called(ctx, account)
	return args.Error(0)
}

func (m *MockAccountRepo) HashPassword(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func (m *MockAccountRepo) VerifyPassword(password, hash string) (bool, error) {
	args := m.Called(password, hash)
	return args.Bool(0), args.Error(1)
}

// MockMessageSender is a mock implementation of MessageSender for testing
type MockMessageSender struct {
	mock.Mock
}

func (m *MockMessageSender) SendSMS(ctx context.Context, phoneNumber, message string) error {
	args := m.Called(ctx, phoneNumber, message)
	return args.Error(0)
}

func (m *MockMessageSender) ValidatePhoneNumber(phoneNumber string) error {
	args := m.Called(phoneNumber)
	return args.Error(0)
}

// MockPhoneNumberVerificationTokenRepo is a mock implementation for testing
type MockPhoneNumberVerificationTokenRepo struct {
	mock.Mock
}

func (m *MockPhoneNumberVerificationTokenRepo) Create(ctx context.Context, phoneNumber string) (string, *PhoneNumberVerificationToken, error) {
	args := m.Called(ctx, phoneNumber)
	return args.String(0), args.Get(1).(*PhoneNumberVerificationToken), args.Error(2)
}

func (m *MockPhoneNumberVerificationTokenRepo) Get(ctx context.Context, verificationToken string) (*PhoneNumberVerificationToken, error) {
	args := m.Called(ctx, verificationToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*PhoneNumberVerificationToken), args.Error(1)
}

func (m *MockPhoneNumberVerificationTokenRepo) GetByPhoneNumber(ctx context.Context, phoneNumber string) (*PhoneNumberVerificationToken, error) {
	args := m.Called(ctx, phoneNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*PhoneNumberVerificationToken), args.Error(1)
}

func (m *MockPhoneNumberVerificationTokenRepo) Delete(ctx context.Context, phoneNumberVerification *PhoneNumberVerificationToken) error {
	args := m.Called(ctx, phoneNumberVerification)
	return args.Error(0)
}

func (m *MockPhoneNumberVerificationTokenRepo) GenerateVerificationToken(length int) (string, error) {
	args := m.Called(length)
	return args.String(0), args.Error(1)
}

func (m *MockPhoneNumberVerificationTokenRepo) HashVerificationToken(token string) string {
	args := m.Called(token)
	return args.String(0)
}

// MockEmailVerificationTokenRepo is a mock implementation for testing
type MockEmailVerificationTokenRepo struct {
	mock.Mock
}

func (m *MockEmailVerificationTokenRepo) Create(ctx context.Context, email string) (string, *EmailVerificationToken, error) {
	args := m.Called(ctx, email)
	return args.String(0), args.Get(1).(*EmailVerificationToken), args.Error(2)
}

func (m *MockEmailVerificationTokenRepo) Get(ctx context.Context, verificationToken string) (*EmailVerificationToken, error) {
	args := m.Called(ctx, verificationToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*EmailVerificationToken), args.Error(1)
}

func (m *MockEmailVerificationTokenRepo) GetByEmail(ctx context.Context, email string) (*EmailVerificationToken, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*EmailVerificationToken), args.Error(1)
}

func (m *MockEmailVerificationTokenRepo) Delete(ctx context.Context, emailVerification *EmailVerificationToken) error {
	args := m.Called(ctx, emailVerification)
	return args.Error(0)
}

func (m *MockEmailVerificationTokenRepo) GenerateVerificationToken(length int) (string, error) {
	args := m.Called(length)
	return args.String(0), args.Error(1)
}

func (m *MockEmailVerificationTokenRepo) HashVerificationToken(token string) string {
	args := m.Called(token)
	return args.String(0)
}

func TestAccountService_GetAccountByPhoneNumber(t *testing.T) {
	tests := []struct {
		name         string
		phoneNumber  string
		setupMocks   func(*MockAccountRepo, *MockMessageSender)
		expectError  bool
		errorMessage string
	}{
		{
			name:        "successful retrieval with valid phone number",
			phoneNumber: "+12345678901",
			setupMocks: func(mockRepo *MockAccountRepo, mockSMS *MockMessageSender) {
				mockAccount := &Account{
					CoreModel:    core.CoreModel{ID: 1},
					Email:        "test@example.com",
					FullName:     "Test User",
					PhoneNumber:  strPtr("+12345678901"),
				}
				mockRepo.On("GetByPhoneNumber", mock.Anything, "+12345678901").Return(mockAccount, nil)
			},
			expectError: false,
		},
		{
			name:        "account not found",
			phoneNumber: "+12345678901",
			setupMocks: func(mockRepo *MockAccountRepo, mockSMS *MockMessageSender) {
				mockRepo.On("GetByPhoneNumber", mock.Anything, "+12345678901").Return(nil, ErrAccountNotFound)
			},
			expectError: true,
		},
		{
			name:        "invalid phone number format",
			phoneNumber: "invalid-phone",
			setupMocks: func(mockRepo *MockAccountRepo, mockSMS *MockMessageSender) {
				// No repo calls should be made for invalid phone numbers
			},
			expectError:  true,
			errorMessage: "invalid phone number format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockRepo := &MockAccountRepo{}
			mockSMS := &MockMessageSender{}
			mockPhoneTokenRepo := &MockPhoneNumberVerificationTokenRepo{}
			mockEmailTokenRepo := &MockEmailVerificationTokenRepo{}

			tt.setupMocks(mockRepo, mockSMS)

			// Create service
			service := NewAccountService(
				mockRepo,
				mockPhoneTokenRepo,
				mockEmailTokenRepo,
				mockSMS,
				nil, // S3 client not needed for this test
				zap.NewNop(),
			)

			// Execute test
			result, err := service.GetAccountByPhoneNumber(context.Background(), tt.phoneNumber)

			// Assertions
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMessage != "" {
					assert.Contains(t, err.Error(), tt.errorMessage)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAccountService_UpdateAccountFullName(t *testing.T) {
	tests := []struct {
		name         string
		accountID    int64
		fullName     string
		setupMocks   func(*MockAccountRepo)
		expectError  bool
		errorMessage string
	}{
		{
			name:      "successful update",
			accountID: 1,
			fullName:  "John Doe",
			setupMocks: func(mockRepo *MockAccountRepo) {
				existingAccount := &Account{
					CoreModel: core.CoreModel{ID: 1},
					Email:     "test@example.com",
					FullName:  "Old Name",
				}
				updatedAccount := &Account{
					CoreModel: core.CoreModel{ID: 1},
					Email:     "test@example.com",
					FullName:  "John Doe",
				}
				mockRepo.On("Get", mock.Anything, int64(1)).Return(existingAccount, nil)
				mockRepo.On("Update", mock.Anything, existingAccount, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(updatedAccount, nil)
			},
			expectError: false,
		},
		{
			name:      "empty full name",
			accountID: 1,
			fullName:  "   ",
			setupMocks: func(mockRepo *MockAccountRepo) {
				// No repo calls should be made for invalid input
			},
			expectError:  true,
			errorMessage: "full name cannot be empty",
		},
		{
			name:      "account not found",
			accountID: 999,
			fullName:  "John Doe",
			setupMocks: func(mockRepo *MockAccountRepo) {
				mockRepo.On("Get", mock.Anything, int64(999)).Return(nil, ErrAccountNotFound)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockRepo := &MockAccountRepo{}
			mockSMS := &MockMessageSender{}
			mockPhoneTokenRepo := &MockPhoneNumberVerificationTokenRepo{}
			mockEmailTokenRepo := &MockEmailVerificationTokenRepo{}

			tt.setupMocks(mockRepo)

			// Create service
			service := NewAccountService(
				mockRepo,
				mockPhoneTokenRepo,
				mockEmailTokenRepo,
				mockSMS,
				nil,
				zap.NewNop(),
			)

			// Execute test
			result, err := service.UpdateAccountFullName(context.Background(), tt.accountID, tt.fullName)

			// Assertions
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMessage != "" {
					assert.Contains(t, err.Error(), tt.errorMessage)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.fullName, result.FullName)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAccountService_UpdateAccountPhoneNumber(t *testing.T) {
	tests := []struct {
		name         string
		accountID    int64
		phoneNumber  string
		setupMocks   func(*MockAccountRepo, *MockMessageSender)
		expectError  bool
		errorMessage string
	}{
		{
			name:        "successful update",
			accountID:   1,
			phoneNumber: "+12345678901",
			setupMocks: func(mockRepo *MockAccountRepo, mockSMS *MockMessageSender) {
				existingAccount := &Account{
					CoreModel: core.CoreModel{ID: 1},
					Email:     "test@example.com",
					FullName:  "Test User",
				}
				updatedAccount := &Account{
					CoreModel:    core.CoreModel{ID: 1},
					Email:        "test@example.com",
					FullName:     "Test User",
					PhoneNumber:  strPtr("+12345678901"),
				}
				mockRepo.On("Get", mock.Anything, int64(1)).Return(existingAccount, nil)
				mockRepo.On("Update", mock.Anything, existingAccount, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(updatedAccount, nil)
			},
			expectError: false,
		},
		{
			name:        "invalid phone number format",
			accountID:   1,
			phoneNumber: "invalid-phone",
			setupMocks: func(mockRepo *MockAccountRepo, mockSMS *MockMessageSender) {
				// No repo calls should be made for invalid phone numbers
			},
			expectError:  true,
			errorMessage: "invalid phone number format",
		},
		{
			name:        "phone number already exists",
			accountID:   1,
			phoneNumber: "+12345678901",
			setupMocks: func(mockRepo *MockAccountRepo, mockSMS *MockMessageSender) {
				existingAccount := &Account{
					CoreModel: core.CoreModel{ID: 1},
					Email:     "test@example.com",
					FullName:  "Test User",
				}
				mockRepo.On("Get", mock.Anything, int64(1)).Return(existingAccount, nil)
				mockRepo.On("Update", mock.Anything, existingAccount, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, ErrPhoneAlreadyExists)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockRepo := &MockAccountRepo{}
			mockSMS := &MockMessageSender{}
			mockPhoneTokenRepo := &MockPhoneNumberVerificationTokenRepo{}
			mockEmailTokenRepo := &MockEmailVerificationTokenRepo{}

			tt.setupMocks(mockRepo, mockSMS)

			// Create service
			service := NewAccountService(
				mockRepo,
				mockPhoneTokenRepo,
				mockEmailTokenRepo,
				mockSMS,
				nil,
				zap.NewNop(),
			)

			// Execute test
			result, err := service.UpdateAccountPhoneNumber(context.Background(), tt.accountID, tt.phoneNumber)

			// Assertions
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMessage != "" {
					assert.Contains(t, err.Error(), tt.errorMessage)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAccountService_validatePhoneNumber(t *testing.T) {
	tests := []struct {
		name        string
		phoneNumber string
		expectError bool
	}{
		{
			name:        "valid US phone number",
			phoneNumber: "+12345678901",
			expectError: false,
		},
		{
			name:        "valid UK phone number",
			phoneNumber: "+447911123456",
			expectError: false,
		},
		{
			name:        "empty phone number",
			phoneNumber: "",
			expectError: true,
		},
		{
			name:        "whitespace only phone number",
			phoneNumber: "   \t\n   ",
			expectError: true,
		},
		{
			name:        "invalid phone number",
			phoneNumber: "123456",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create service with minimal dependencies
			service := NewAccountService(
				&MockAccountRepo{},
				&MockPhoneNumberVerificationTokenRepo{},
				&MockEmailVerificationTokenRepo{},
				&MockMessageSender{},
				nil,
				zap.NewNop(),
			)

			err := service.validatePhoneNumber(tt.phoneNumber)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper function
func strPtr(s string) *string {
	return &s
}