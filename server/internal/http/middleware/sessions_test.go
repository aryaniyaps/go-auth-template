package httpmiddleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestSessionMiddleware_EncryptAndDecryptSession(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Create a test configuration with exactly 16 bytes for AES-128
	config := SessionConfig{
		JWESecretKey:  "test-secret-16b!", // exactly 16 bytes
		SessionCookie: "session",
		MaxAge:        3600,
		Path:          "/",
		SameSite:      http.SameSiteLaxMode,
		Secure:        false,
		Domain:        "",
	}

	middleware := NewSessionMiddleware(config, logger)

	// Create a test handler that sets session data
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get session data from context
		sessionData := r.Context().Value("session_data").(map[string]interface{})

		// Set some session data
		sessionData["user_id"] = "12345"
		sessionData["email"] = "test@example.com"
		sessionData["role"] = "user"

		// Update context
		ctx := context.WithValue(r.Context(), "session_data", sessionData)
		*r = *r.WithContext(ctx)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	// First request to set session
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())

	// Check that session cookie was set
	cookies := w.Result().Cookies()
	assert.Len(t, cookies, 1)
	assert.Equal(t, "session", cookies[0].Name)
	assert.True(t, cookies[0].HttpOnly)
	assert.Equal(t, "/", cookies[0].Path)
	assert.Equal(t, 3600, cookies[0].MaxAge)

	// Second request to retrieve session data
	req2 := httptest.NewRequest("GET", "/", nil)
	req2.Header = make(http.Header)
	req2.Header.Add("Cookie", cookies[0].String())

	w2 := httptest.NewRecorder()
	handler2 := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get session data from context
		sessionData := r.Context().Value("session_data").(map[string]interface{})

		// Verify session data was decrypted correctly
		assert.Equal(t, "12345", sessionData["user_id"])
		assert.Equal(t, "test@example.com", sessionData["email"])
		assert.Equal(t, "user", sessionData["role"])

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Session retrieved"))
	}))

	handler2.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)
	assert.Equal(t, "Session retrieved", w2.Body.String())
}

func TestSessionMiddleware_MissingCookie(t *testing.T) {
	logger := zaptest.NewLogger(t)

	config := SessionConfig{
		JWESecretKey:  "test-secret-16b!", // exactly 16 bytes
		SessionCookie: "session",
		MaxAge:        3600,
		Path:          "/",
		SameSite:      http.SameSiteLaxMode,
		Secure:        false,
		Domain:        "",
	}

	middleware := NewSessionMiddleware(config, logger)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check context values exist but are empty
		sessionData := r.Context().Value("session_data").(map[string]interface{})
		assert.Empty(t, sessionData)

		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// No cookies should be set
	cookies := w.Result().Cookies()
	assert.Len(t, cookies, 0)
}

func TestSessionMiddleware_InvalidJWE(t *testing.T) {
	logger := zaptest.NewLogger(t)

	config := SessionConfig{
		JWESecretKey:  "test-secret-16b!", // exactly 16 bytes
		SessionCookie: "session",
		MaxAge:        3600,
		Path:          "/",
		SameSite:      http.SameSiteLaxMode,
		Secure:        false,
		Domain:        "",
	}

	middleware := NewSessionMiddleware(config, logger)

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  "session",
		Value: "invalid-jwe-token",
	})

	w := httptest.NewRecorder()

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check context values exist but are empty (due to JWE decryption failure)
		sessionData := r.Context().Value("session_data").(map[string]interface{})
		assert.Empty(t, sessionData)

		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSessionMiddleware_SessionCleared(t *testing.T) {
	logger := zaptest.NewLogger(t)

	config := SessionConfig{
		JWESecretKey:  "test-secret-16b!", // exactly 16 bytes
		SessionCookie: "session",
		MaxAge:        3600,
		Path:          "/",
		SameSite:      http.SameSiteLaxMode,
		Secure:        false,
		Domain:        "",
	}

	middleware := NewSessionMiddleware(config, logger)

	// First, create a session by setting some data
	setHandler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionData := r.Context().Value("session_data").(map[string]interface{})
		sessionData["user_id"] = "12345"

		ctx := context.WithValue(r.Context(), "session_data", sessionData)
		*r = *r.WithContext(ctx)

		w.WriteHeader(http.StatusOK)
	}))

	req1 := httptest.NewRequest("GET", "/", nil)
	w1 := httptest.NewRecorder()
	setHandler.ServeHTTP(w1, req1)

	// Should have a session cookie
	cookies1 := w1.Result().Cookies()
	assert.Len(t, cookies1, 1)

	// Now clear the session
	clearHandler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Clear session data
		ctx := context.WithValue(r.Context(), "session_data", map[string]interface{}{})
		*r = *r.WithContext(ctx)

		w.WriteHeader(http.StatusOK)
	}))

	req2 := httptest.NewRequest("GET", "/", nil)
	req2.Header = make(http.Header)
	req2.Header.Add("Cookie", cookies1[0].String())

	w2 := httptest.NewRecorder()
	clearHandler.ServeHTTP(w2, req2)

	// Should have a deletion cookie
	cookies2 := w2.Result().Cookies()
	assert.Len(t, cookies2, 1)
	assert.Equal(t, "session", cookies2[0].Name)
	assert.Equal(t, -1, cookies2[0].MaxAge) // MaxAge -1 indicates cookie deletion
}