package resolver

import (
	"server/internal/infrastructure/captcha"
)

type Resolver struct {
	// add services here
	// UserService *services.UserService
	captchaVerifier captcha.CaptchaVerifier
}

// constructor for Fx
func NewResolver(captchaVerifier captcha.CaptchaVerifier) *Resolver {
	return &Resolver{
		captchaVerifier: captchaVerifier,
	}
}
