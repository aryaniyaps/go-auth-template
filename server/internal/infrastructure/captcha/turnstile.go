package captcha

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// TurnstileVerifier implements BaseCaptchaVerifier using Cloudflare Turnstile.
type TurnstileVerifier struct {
	secretKey string
	siteKey   string
	client    *http.Client
	baseURL   string
}

// turnstileVerifyRequest represents the request payload to Cloudflare Turnstile API.
type turnstileVerifyRequest struct {
	Secret   string `json:"secret"`
	Response string `json:"response"`
	RemoteIP string `json:"remoteip,omitempty"`
}

// turnstileVerifyResponse represents the response from Cloudflare Turnstile API.
type turnstileVerifyResponse struct {
	Success     bool     `json:"success"`
	ErrorCodes  []string `json:"error-codes"`
	ChallengeTs string   `json:"challenge_ts"`
	Hostname    string   `json:"hostname"`
	Action      string   `json:"action,omitempty"`
	Cdata       string   `json:"cdata,omitempty"`
}

// NewTurnstileVerifier creates a new Cloudflare Turnstile verifier.
func NewTurnstileVerifier(secretKey, siteKey string) *TurnstileVerifier {
	return &TurnstileVerifier{
		secretKey: secretKey,
		siteKey:   siteKey,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://challenges.cloudflare.com/turnstile/v0/siteverify",
	}
}

// VerifyToken implements BaseCaptchaVerifier interface by validating the token
// with Cloudflare Turnstile API.
func (tv *TurnstileVerifier) VerifyToken(ctx context.Context, captchaToken string) (bool, error) {
	if captchaToken == "" {
		return false, ErrEmptyToken
	}

	if tv.secretKey == "" {
		return false, ErrMissingSecretKey
	}

	// Prepare the request payload
	requestData := turnstileVerifyRequest{
		Secret:   tv.secretKey,
		Response: captchaToken,
	}

	// Convert request to JSON
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return false, fmt.Errorf("failed to marshal request data: %w", err)
	}

	// Create HTTP request with context
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tv.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "go-auth-template/1.0")

	// Send request
	resp, err := tv.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse response
	var verifyResp turnstileVerifyResponse
	if err := json.Unmarshal(body, &verifyResp); err != nil {
		return false, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check HTTP status code
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("turnstile API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Check verification result
	if !verifyResp.Success {
		// Map specific error codes to our custom errors
		if len(verifyResp.ErrorCodes) > 0 {
			for _, errorCode := range verifyResp.ErrorCodes {
				switch errorCode {
				case "missing-input-secret":
					return false, ErrMissingSecretKey
				case "invalid-input-secret":
					return false, ErrInvalidSecretKey
				case "missing-input-response":
					return false, ErrEmptyToken
				case "invalid-input-response":
					return false, ErrInvalidToken
				case "bad-request":
					return false, ErrBadRequest
				case "timeout-or-duplicate":
					return false, ErrExpiredToken
				}
			}
		}
		return false, ErrInvalidToken
	}

	return true, nil
}

// String returns a string representation of the verifier.
func (tv *TurnstileVerifier) String() string {
	return fmt.Sprintf("TurnstileVerifier{siteKey: %s, baseURL: %s}", tv.siteKey, tv.baseURL)
}
