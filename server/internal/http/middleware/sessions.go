package httpmiddleware

import (
	"context"
	"net/http"

	"github.com/go-chi/jwtauth/v5"
	"go.uber.org/zap"
)

// SessionMiddleware provides session middleware for injecting session token data into context
type SessionMiddleware struct {
	jwtAuth *jwtauth.JWTAuth
	logger  *zap.Logger
}

// NewSessionMiddleware creates a new session middleware
func NewSessionMiddleware(jwtAuth *jwtauth.JWTAuth, logger *zap.Logger) func(http.Handler) http.Handler {
	middleware := &SessionMiddleware{
		jwtAuth: jwtAuth,
		logger:  logger,
	}

	return middleware.Handler
}

// Handler is the middleware function that handles session token data
func (sm *SessionMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Try to get token from cookie
		sessionToken := sm.extractTokenFromCookie(r)

		var tokenData map[string]interface{}

		// If we have a session token, parse it and extract data
		if sessionToken != "" {
			tokenData = sm.parseTokenData(sessionToken)
		}

		// Inject session token data into context
		ctx = sm.injectSessionContext(ctx, tokenData)

		// Continue to next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// extractTokenFromCookie extracts the session token from the HTTP-only cookie
func (sm *SessionMiddleware) extractTokenFromCookie(r *http.Request) string {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		// Cookie not found or error reading it
		return ""
	}

	return cookie.Value
}

// parseTokenData parses JWT token and extracts claims data without database calls
func (sm *SessionMiddleware) parseTokenData(sessionToken string) map[string]interface{} {
	// Validate JWT token structure and extract claims
	token, err := jwtauth.VerifyToken(sm.jwtAuth, sessionToken)
	if err != nil {
		sm.logger.Debug("Invalid JWT token", zap.Error(err))
		return nil
	}

	// Extract claims from token
	claims, err := token.AsMap(context.Background())
	if err != nil {
		sm.logger.Debug("Failed to extract token claims", zap.Error(err))
		return nil
	}

	return claims
}

// injectSessionContext injects session token data into the request context
func (sm *SessionMiddleware) injectSessionContext(ctx context.Context, tokenData map[string]interface{}) context.Context {
	// Use context key "session_token_data" for the parsed token data
	ctx = context.WithValue(ctx, "session_token_data", tokenData)

	return ctx
}