package logger

import (
	"context"
	"server/internal/config"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(lc fx.Lifecycle, cfg *config.Config) *zap.Logger {
	logger := zap.Must(zap.NewProduction())
	if cfg.Environment == "development" {
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		logger = zap.Must(config.Build())

	}

	if cfg.Environment == "testing" {
		logger = zap.NewNop()
	}

	lc.Append(fx.Hook{
		OnStop: func(context.Context) error {
			return logger.Sync()
		},
	})
	return logger
}
