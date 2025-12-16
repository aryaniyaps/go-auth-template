package captcha

import "context"

// BaseCaptchaVerifier defines the interface for captcha verification services.
// Implementations should validate captcha tokens and return whether verification
// was successful along with any error information.
type BaseCaptchaVerifier interface {
	// VerifyToken validates a captcha token using context for cancellation and timeout
	// Returns true if the token is valid, false otherwise
	// Returns an error if verification cannot be completed
	VerifyToken(ctx context.Context, captchaToken string) (bool, error)
}
