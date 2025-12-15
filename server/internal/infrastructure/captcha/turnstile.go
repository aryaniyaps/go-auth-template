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

// TurnstileResponse represents the response from Cloudflare Turnstile API
type TurnstileResponse struct {
	Success    bool     `json:"success"`
	ErrorCodes []string `json:"error-codes"`
	ChallengeTS string  `json:"challenge_ts"`
	Hostname    string  `json:"hostname"`
}

// TurnstileVerifier implements captcha verification using Cloudflare Turnstile
type TurnstileVerifier struct {
	secretKey string
	client    *http.Client
}

// NewTurnstileVerifier creates a new Turnstile verifier
func NewTurnstileVerifier(config *CaptchaConfig) (CaptchaVerifier, error) {
	if config.CloudflareSecretKey == "" {
		return nil, fmt.Errorf("%w: Cloudflare secret key is required", ErrProviderNotConfigured)
	}

	return &TurnstileVerifier{
		secretKey: config.CloudflareSecretKey,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

// VerifyToken verifies a Turnstile token with Cloudflare's API
func (t *TurnstileVerifier) VerifyToken(ctx context.Context, token string) (bool, error) {
	if token == "" {
		return false, ErrEmptyToken
	}

	// Create the request data
	data := map[string]string{
		"secret": t.secretKey,
		"response": token,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return false, fmt.Errorf("%w: failed to marshal request data", ErrVerificationFailed)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", "https://challenges.cloudflare.com/turnstile/v0/siteverify", bytes.NewBuffer(jsonData))
	if err != nil {
		return false, fmt.Errorf("%w: failed to create request", ErrVerificationFailed)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := t.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("%w: failed to send request", ErrVerificationFailed)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("%w: failed to read response", ErrVerificationFailed)
	}

	// Parse response
	var turnstileResp TurnstileResponse
	if err := json.Unmarshal(body, &turnstileResp); err != nil {
		return false, fmt.Errorf("%w: failed to parse response", ErrVerificationFailed)
	}

	return turnstileResp.Success, nil
}