package http

import (
	"server/internal/config"

	"github.com/go-chi/jwtauth/v5"
	"go.uber.org/fx"
)

// HTTPTopLevelModule contains HTTP-level dependencies
var HTTPTopLevelModule = fx.Options(
	fx.Provide(
		NewJWTAuth,
	),
)

// NewJWTAuth creates a new JWT authentication instance
func NewJWTAuth(cfg *config.Config) *jwtauth.JWTAuth {
	return jwtauth.New("HS256", []byte(cfg.JWTSecret), nil)
}