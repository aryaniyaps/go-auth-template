package http

import (
	"context"
	"net"
	"net/http"
	"server/internal/config"
	httpmiddleware "server/internal/http/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func addMiddleware(r *chi.Mux, log *zap.Logger) {
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(httpmiddleware.LoggerMiddleware(log))
	r.Use(middleware.AllowContentType("application/json"))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowCredentials: true,
	}))
	r.Use(middleware.Heartbeat("/ping"))
	r.Use(middleware.Recoverer)
}

func New(lc fx.Lifecycle, cfg *config.Config, log *zap.Logger) *chi.Mux {
	r := chi.NewRouter()
	addMiddleware(r, log)

	srv := &http.Server{Addr: ":" + cfg.ServerPort, Handler: r}

	if cfg.Environment != "testing" {
		// don't start server in testing environment
		lc.Append(fx.Hook{
			OnStart: func(context.Context) error {
				ln, err := net.Listen("tcp", srv.Addr)
				if err != nil {
					return err
				}
				log.Info("🚀 Starting HTTP server", zap.String("addr", srv.Addr))
				go srv.Serve(ln)
				return nil
			},
			OnStop: func(ctx context.Context) error {
				return srv.Shutdown(ctx)
			},
		})
	}

	return r
}
