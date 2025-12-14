package httpmiddleware

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-jose/go-jose/v4"
	"go.uber.org/zap"
)

// SessionMiddleware provides JWE-based session middleware for encrypting/decrypting session data
type SessionMiddleware struct {
	jweSecretKey  []byte
	logger        *zap.Logger
	sessionCookie string
	maxAge        int
	path          string
	sameSite      http.SameSite
	secure        bool
	domain        string
}

// SessionConfig holds configuration for the session middleware
type SessionConfig struct {
	JWESecretKey  string
	SessionCookie string
	MaxAge        int
	Path          string
	SameSite      http.SameSite
	Secure        bool
	Domain        string
}

// NewSessionMiddleware creates a new JWE-based session middleware
func NewSessionMiddleware(config SessionConfig, logger *zap.Logger) func(http.Handler) http.Handler {
	middleware := &SessionMiddleware{
		jweSecretKey:  []byte(config.JWESecretKey),
		logger:        logger,
		sessionCookie: config.SessionCookie,
		maxAge:        config.MaxAge,
		path:          config.Path,
		sameSite:      config.SameSite,
		secure:        config.Secure,
		domain:        config.Domain,
	}

	return middleware.Handler
}

// Handler is the middleware function that handles JWE session data
func (sm *SessionMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Try to get the session from the cookie
		sessionData := sm.extractSessionFromCookie(r)

		// Inject session data into context
		ctx = context.WithValue(ctx, "session_data", sessionData)

		// Create a response writer wrapper to capture session modifications
		wrapper := &sessionResponseWriter{
			ResponseWriter: w,
			request:        r,
			middleware:     sm,
			sessionData:    sessionData,
		}

		// Continue to next handler with our custom response writer
		next.ServeHTTP(wrapper, r.WithContext(ctx))
	})
}

// extractSessionFromCookie extracts and decrypts the JWE session from the HTTP-only cookie
func (sm *SessionMiddleware) extractSessionFromCookie(r *http.Request) map[string]interface{} {
	cookie, err := r.Cookie(sm.sessionCookie)
	if err != nil {
		// Cookie not found or error reading it
		return make(map[string]interface{})
	}

	// Parse the JWE object
	parsedJwe, err := jose.ParseEncrypted(cookie.Value, []jose.KeyAlgorithm{jose.A128GCMKW}, []jose.ContentEncryption{jose.A128GCM})
	if err != nil {
		sm.logger.Debug("Failed to parse JWE from cookie", zap.Error(err))
		return make(map[string]interface{})
	}

	// Decrypt the JWE to get session data
	sessionDataBytes, err := parsedJwe.Decrypt(sm.jweSecretKey)
	if err != nil {
		sm.logger.Debug("Failed to decrypt JWE session", zap.Error(err))
		return make(map[string]interface{})
	}

	// Parse JSON session data
	var sessionData map[string]interface{}
	if err := json.Unmarshal(sessionDataBytes, &sessionData); err != nil {
		sm.logger.Debug("Failed to unmarshal session data", zap.Error(err))
		return make(map[string]interface{})
	}

	return sessionData
}

// encryptSession encrypts session data into a JWE token
func (sm *SessionMiddleware) encryptSession(sessionData map[string]interface{}) (string, error) {
	// Use AES-128-GCM with key wrapping for better key size flexibility
	encrypter, err := jose.NewEncrypter(
		jose.A128GCM,
		jose.Recipient{Algorithm: jose.A128GCMKW, Key: sm.jweSecretKey},
		&jose.EncrypterOptions{},
	)
	if err != nil {
		return "", err
	}

	// Marshal session data to JSON
	sessionDataBytes, err := json.Marshal(sessionData)
	if err != nil {
		return "", err
	}

	// Encrypt the session data
	jweObject, err := encrypter.Encrypt(sessionDataBytes)
	if err != nil {
		return "", err
	}

	// Serialize to compact format
	return jweObject.CompactSerialize()
}

// sessionResponseWriter wraps http.ResponseWriter to handle session cookie operations
type sessionResponseWriter struct {
	http.ResponseWriter
	request     *http.Request
	middleware  *SessionMiddleware
	sessionData map[string]interface{}
	written     bool
}

// WriteHeader intercepts the response headers to set session cookies
func (srw *sessionResponseWriter) WriteHeader(statusCode int) {
	if !srw.written {
		srw.handleSessionCookie()
		srw.written = true
	}
	srw.ResponseWriter.WriteHeader(statusCode)
}

// Write intercepts the response body to ensure session cookies are set
func (srw *sessionResponseWriter) Write(data []byte) (int, error) {
	if !srw.written {
		srw.handleSessionCookie()
		srw.written = true
	}
	return srw.ResponseWriter.Write(data)
}

// handleSessionCookie sets or deletes the session cookie based on session data
func (srw *sessionResponseWriter) handleSessionCookie() {
	// Get updated session data from context if available
	var currentSessionData map[string]interface{}
	if ctx := srw.request.Context(); ctx != nil {
		if sessionData, ok := ctx.Value("session_data").(map[string]interface{}); ok {
			currentSessionData = sessionData
		}
	}

	// If session data exists, encrypt and set the cookie
	if len(currentSessionData) > 0 {
		token, err := srw.middleware.encryptSession(currentSessionData)
		if err != nil {
			srw.middleware.logger.Error("Failed to encrypt session", zap.Error(err))
			return
		}

		cookie := &http.Cookie{
			Name:     srw.middleware.sessionCookie,
			Value:    token,
			MaxAge:   srw.middleware.maxAge,
			Path:     srw.middleware.path,
			HttpOnly: true,
			Secure:   srw.middleware.secure,
			SameSite: srw.middleware.sameSite,
			Domain:   srw.middleware.domain,
		}
		http.SetCookie(srw.ResponseWriter, cookie)
	} else if len(srw.sessionData) > 0 && len(currentSessionData) == 0 {
		// If session was cleared during the request (initially had data, now empty), delete the cookie
		cookie := &http.Cookie{
			Name:     srw.middleware.sessionCookie,
			MaxAge:   -1,
			Path:     srw.middleware.path,
			HttpOnly: true,
			Secure:   srw.middleware.secure,
			SameSite: srw.middleware.sameSite,
			Domain:   srw.middleware.domain,
		}
		http.SetCookie(srw.ResponseWriter, cookie)
	}
}
