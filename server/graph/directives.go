package graph

import (
	"context"
	"errors"
	"time"

	"github.com/99designs/gqlgen/graphql"
)

// Authentication errors
var (
	ErrNotAuthenticated = errors.New("User is not authenticated")
	ErrRequiresSudoMode = errors.New("Action requires sudo mode")
)

// IsAuthenticated directive protects fields to ensure only authenticated users can access them
// Based on Python IsAuthenticated permission class
func IsAuthenticated(ctx context.Context, obj interface{}, next graphql.Resolver) (interface{}, error) {
	// Get session token data from context (from SessionMiddleware)
	sessionTokenData := ctx.Value("session_token_data")
	if sessionTokenData == nil {
		return nil, ErrNotAuthenticated
	}

	// Check if session contains user information
	tokenData, ok := sessionTokenData.(map[string]interface{})
	if !ok {
		return nil, ErrNotAuthenticated
	}

	// Verify user_id exists in token data
	if _, exists := tokenData["user_id"]; !exists {
		return nil, ErrNotAuthenticated
	}

	return next(ctx)
}

// RequiresSudoMode directive protects fields that require elevated privileges
// Based on Python RequiresSudoMode permission class
func RequiresSudoMode(ctx context.Context, obj interface{}, next graphql.Resolver) (interface{}, error) {
	// First ensure user is authenticated
	_, err := IsAuthenticated(ctx, obj, func(c context.Context) (interface{}, error) {
		return nil, nil // Just for authentication check
	})
	if err != nil {
		return nil, err
	}

	// Get session token data to check sudo mode
	sessionTokenData := ctx.Value("session_token_data")
	if sessionTokenData == nil {
		return nil, ErrRequiresSudoMode
	}

	tokenData, ok := sessionTokenData.(map[string]interface{})
	if !ok {
		return nil, ErrRequiresSudoMode
	}

	// Check for sudo_mode_expires_at in session data
	sudoModeExpiresAtRaw, exists := tokenData["sudo_mode_expires_at"]
	if !exists {
		return nil, ErrRequiresSudoMode
	}

	// Parse the expiration time
	sudoModeExpiresAtStr, ok := sudoModeExpiresAtRaw.(string)
	if !ok {
		return nil, ErrRequiresSudoMode
	}

	sudoModeExpiresAt, err := time.Parse(time.RFC3339, sudoModeExpiresAtStr)
	if err != nil {
		return nil, ErrRequiresSudoMode
	}

	// Check if sudo mode is still valid
	if time.Now().UTC().After(sudoModeExpiresAt) {
		return nil, ErrRequiresSudoMode
	}

	return next(ctx)
}