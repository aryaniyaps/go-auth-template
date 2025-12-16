package resolver

import (
	"server/internal/infrastructure/captcha"
)

type Resolver struct {
	// add services here
	// UserService *services.UserService
	captchaVerifier captcha.BaseCaptchaVerifier
}

// constructor for Fx
func NewResolver(captchaVerifier captcha.BaseCaptchaVerifier) *Resolver {
	return &Resolver{
		captchaVerifier: captchaVerifier,
	}
}
