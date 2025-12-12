package db

import (
	"context"
	"database/sql"
	"server/internal/config"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
	"go.uber.org/fx"
)

func NewDB(cfg *config.Config, lc fx.Lifecycle) *bun.DB {
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(cfg.DBUrl)))

	// Configure connection pool
	sqldb.SetMaxOpenConns(25)                 // Maximum open connections
	sqldb.SetMaxIdleConns(10)                 // Maximum idle connections
	sqldb.SetConnMaxLifetime(5 * time.Minute) // Connection lifetime
	sqldb.SetConnMaxIdleTime(5 * time.Minute) // Idle connection timeout

	db := bun.NewDB(sqldb, pgdialect.New())

	if cfg.Environment == "development" {
		db.AddQueryHook(bundebug.NewQueryHook(
			bundebug.WithVerbose(true),
		))
	}

	// Add lifecycle hooks for graceful DB connection management
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Verify connection is ready (non-blocking ping)
			return db.Ping()
		},
		OnStop: func(ctx context.Context) error {
			// Graceful shutdown with context timeout
			return db.Close()
		},
	})

	return db
}
