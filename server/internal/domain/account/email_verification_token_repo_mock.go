package account

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockEmailVerificationTokenRepo is a mock implementation of EmailVerificationTokenRepo for testing
type MockEmailVerificationTokenRepo struct {
	mock.Mock
}

func (m *MockEmailVerificationTokenRepo) Create(ctx context.Context, email string) (string, *EmailVerificationToken, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return "", nil, args.Error(1)
	}
	if args.Get(1) == nil {
		return args.String(0), nil, args.Error(1)
	}
	return args.String(0), args.Get(1).(*EmailVerificationToken), args.Error(1)
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