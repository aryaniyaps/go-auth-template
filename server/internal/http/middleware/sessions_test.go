package httpmiddleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/jwtauth/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestSessionMiddleware_ValidSessionToken(t *testing.T) {
	logger := zaptest.NewLogger(t)
	jwtAuth := jwtauth.New("HS256", []byte("test-secret"), nil)

	// Generate valid JWT token with user data
	_, tokenString, _ := jwtAuth.Encode(map[string]interface{}{
		"user_id": 123,
		"email":   "test@example.com",
		"role":    "user",
	})

	middleware := NewSessionMiddleware(jwtAuth, logger)

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  "session_token",
		Value: tokenString,
	})

	w := httptest.NewRecorder()
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check context values
		ctxTokenData := r.Context().Value("session_token_data")

		assert.NotNil(t, ctxTokenData)

		// Verify token data contains expected claims
		tokenData := ctxTokenData.(map[string]interface{})
		assert.Equal(t, float64(123), tokenData["user_id"])
		assert.Equal(t, "test@example.com", tokenData["email"])
		assert.Equal(t, "user", tokenData["role"])

		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSessionMiddleware_MissingCookie(t *testing.T) {
	logger := zaptest.NewLogger(t)
	jwtAuth := jwtauth.New("HS256", []byte("test-secret"), nil)

	middleware := NewSessionMiddleware(jwtAuth, logger)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check context values are nil
		ctxTokenData := r.Context().Value("session_token_data")

		assert.Nil(t, ctxTokenData)

		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSessionMiddleware_InvalidToken(t *testing.T) {
	logger := zaptest.NewLogger(t)
	jwtAuth := jwtauth.New("HS256", []byte("test-secret"), nil)

	middleware := NewSessionMiddleware(jwtAuth, logger)

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  "session_token",
		Value: "invalid-token-string",
	})

	w := httptest.NewRecorder()

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check context values are nil
		ctxTokenData := r.Context().Value("session_token_data")

		assert.Nil(t, ctxTokenData)

		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSessionMiddleware_TokenWithoutClaims(t *testing.T) {
	logger := zaptest.NewLogger(t)
	jwtAuth := jwtauth.New("HS256", []byte("test-secret"), nil)

	// Generate JWT token with no claims
	_, tokenString, _ := jwtAuth.Encode(map[string]interface{}{})

	middleware := NewSessionMiddleware(jwtAuth, logger)

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  "session_token",
		Value: tokenString,
	})

	w := httptest.NewRecorder()

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check context values exist but are empty
		ctxTokenData := r.Context().Value("session_token_data")

		assert.NotNil(t, ctxTokenData)

		// Verify token data is empty map
		tokenData := ctxTokenData.(map[string]interface{})
		assert.Empty(t, tokenData)

		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSessionMiddleware_RequestAlwaysContinues(t *testing.T) {
	logger := zaptest.NewLogger(t)
	jwtAuth := jwtauth.New("HS256", []byte("test-secret"), nil)

	middleware := NewSessionMiddleware(jwtAuth, logger)

	// Test with various token states
	testCases := []struct {
		name   string
		token  string
		setup  func(*http.Request)
	}{
		{
			name:  "no_cookie",
			token: "",
		},
		{
			name:  "invalid_token",
			token: "invalid-jwt",
			setup: func(req *http.Request) {
				req.AddCookie(&http.Cookie{
					Name:  "session_token",
					Value: "invalid-jwt",
				})
			},
		},
		{
			name:  "valid_token",
			token: "valid-token",
			setup: func(req *http.Request) {
				_, tokenString, _ := jwtAuth.Encode(map[string]interface{}{"user_id": 123})
				req.AddCookie(&http.Cookie{
					Name:  "session_token",
					Value: tokenString,
				})
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			if tc.setup != nil {
				tc.setup(req)
			}

			w := httptest.NewRecorder()

			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Request should always continue normally
				w.WriteHeader(http.StatusOK)
			}))

			handler.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}