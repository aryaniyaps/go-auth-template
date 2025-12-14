package graph

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockResolver is a mock implementation of a GraphQL resolver
type MockResolver struct {
	mock.Mock
}

func (m *MockResolver) Resolve(ctx context.Context) (interface{}, error) {
	args := m.Called(ctx)
	return args.Get(0), args.Error(1)
}

func TestIsAuthenticated_WithValidSessionToken(t *testing.T) {
	// Create session token data with user_id
	tokenData := map[string]interface{}{
		"user_id": float64(123),
		"email":   "test@example.com",
	}

	// Create context with session token data
	ctx := context.WithValue(context.Background(), "session_token_data", tokenData)

	// Mock resolver
	mockResolver := &MockResolver{}
	mockResolver.On("Resolve", ctx).Return("success", nil)

	// Test directive
	result, err := IsAuthenticated(ctx, nil, mockResolver.Resolve)

	assert.NoError(t, err)
	assert.Equal(t, "success", result)
	mockResolver.AssertExpectations(t)
}

func TestIsAuthenticated_WithoutSessionTokenData(t *testing.T) {
	// Create context without session token data
	ctx := context.Background()

	// Mock resolver
	mockResolver := &MockResolver{}

	// Test directive
	result, err := IsAuthenticated(ctx, nil, mockResolver.Resolve)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrNotAuthenticated)
	mockResolver.AssertNotCalled(t, "Resolve")
}

func TestIsAuthenticated_WithNilSessionTokenData(t *testing.T) {
	// Create context with nil session token data
	ctx := context.WithValue(context.Background(), "session_token_data", nil)

	// Mock resolver
	mockResolver := &MockResolver{}

	// Test directive
	result, err := IsAuthenticated(ctx, nil, mockResolver.Resolve)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrNotAuthenticated)
	mockResolver.AssertNotCalled(t, "Resolve")
}

func TestIsAuthenticated_WithInvalidSessionTokenData(t *testing.T) {
	// Create context with invalid session token data type
	ctx := context.WithValue(context.Background(), "session_token_data", "invalid-data")

	// Mock resolver
	mockResolver := &MockResolver{}

	// Test directive
	result, err := IsAuthenticated(ctx, nil, mockResolver.Resolve)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrNotAuthenticated)
	mockResolver.AssertNotCalled(t, "Resolve")
}

func TestIsAuthenticated_WithoutUserID(t *testing.T) {
	// Create session token data without user_id
	tokenData := map[string]interface{}{
		"email": "test@example.com",
	}

	// Create context with session token data
	ctx := context.WithValue(context.Background(), "session_token_data", tokenData)

	// Mock resolver
	mockResolver := &MockResolver{}

	// Test directive
	result, err := IsAuthenticated(ctx, nil, mockResolver.Resolve)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrNotAuthenticated)
	mockResolver.AssertNotCalled(t, "Resolve")
}

func TestRequiresSudoMode_WithValidSudoMode(t *testing.T) {
	// Create session token data with valid sudo mode
	futureTime := time.Now().UTC().Add(15 * time.Minute)
	tokenData := map[string]interface{}{
		"user_id":              float64(123),
		"sudo_mode_expires_at": futureTime.Format(time.RFC3339),
	}

	// Create context with session token data
	ctx := context.WithValue(context.Background(), "session_token_data", tokenData)

	// Mock resolver
	mockResolver := &MockResolver{}
	mockResolver.On("Resolve", ctx).Return("sudo-success", nil)

	// Test directive
	result, err := RequiresSudoMode(ctx, nil, mockResolver.Resolve)

	assert.NoError(t, err)
	assert.Equal(t, "sudo-success", result)
	mockResolver.AssertExpectations(t)
}

func TestRequiresSudoMode_WithExpiredSudoMode(t *testing.T) {
	// Create session token data with expired sudo mode
	pastTime := time.Now().UTC().Add(-15 * time.Minute)
	tokenData := map[string]interface{}{
		"user_id":              float64(123),
		"sudo_mode_expires_at": pastTime.Format(time.RFC3339),
	}

	// Create context with session token data
	ctx := context.WithValue(context.Background(), "session_token_data", tokenData)

	// Mock resolver
	mockResolver := &MockResolver{}

	// Test directive
	result, err := RequiresSudoMode(ctx, nil, mockResolver.Resolve)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrRequiresSudoMode)
	mockResolver.AssertNotCalled(t, "Resolve")
}

func TestRequiresSudoMode_WithoutSudoMode(t *testing.T) {
	// Create session token data without sudo mode
	tokenData := map[string]interface{}{
		"user_id": float64(123),
	}

	// Create context with session token data
	ctx := context.WithValue(context.Background(), "session_token_data", tokenData)

	// Mock resolver
	mockResolver := &MockResolver{}

	// Test directive
	result, err := RequiresSudoMode(ctx, nil, mockResolver.Resolve)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrRequiresSudoMode)
	mockResolver.AssertNotCalled(t, "Resolve")
}

func TestRequiresSudoMode_WithInvalidSudoExpiryFormat(t *testing.T) {
	// Create session token data with invalid sudo mode expiry format
	tokenData := map[string]interface{}{
		"user_id":              float64(123),
		"sudo_mode_expires_at": "invalid-timestamp",
	}

	// Create context with session token data
	ctx := context.WithValue(context.Background(), "session_token_data", tokenData)

	// Mock resolver
	mockResolver := &MockResolver{}

	// Test directive
	result, err := RequiresSudoMode(ctx, nil, mockResolver.Resolve)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrRequiresSudoMode)
	mockResolver.AssertNotCalled(t, "Resolve")
}

func TestRequiresSudoMode_WithoutAuthentication(t *testing.T) {
	// Create context without session token data
	ctx := context.Background()

	// Mock resolver
	mockResolver := &MockResolver{}

	// Test directive
	result, err := RequiresSudoMode(ctx, nil, mockResolver.Resolve)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrNotAuthenticated)
	mockResolver.AssertNotCalled(t, "Resolve")
}

func TestRequiresSudoMode_CallsIsAuthenticatedFirst(t *testing.T) {
	// This test verifies that RequiresSudoMode calls IsAuthenticated first
	// by creating a scenario where authentication fails

	// Create context without session token data
	ctx := context.Background()

	// Mock resolver
	mockResolver := &MockResolver{}

	// Test directive
	result, err := RequiresSudoMode(ctx, nil, mockResolver.Resolve)

	// Should return authentication error, not sudo mode error
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrNotAuthenticated)
	mockResolver.AssertNotCalled(t, "Resolve")
}