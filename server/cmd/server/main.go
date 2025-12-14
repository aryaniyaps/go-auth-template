//go:generate go run github.com/99designs/gqlgen generate

package main

import (
	"server/graph"
	"server/graph/generated"
	"server/graph/resolver"
	"server/internal/config"
	"server/internal/domain/account"
	"server/internal/domain/auth"
	serverhttp "server/internal/http"
	"server/internal/infrastructure/db"
	"server/internal/infrastructure/s3client"
	"server/internal/logger"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi/v5"
	"github.com/vektah/gqlparser/v2/ast"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func AddGraphQLHandler(r *chi.Mux, cfg *config.Config) {
	resolver := &resolver.Resolver{}

	srv := handler.New(generated.NewExecutableSchema(generated.Config{
		Resolvers: resolver,
		Directives: generated.DirectiveRoot{
			IsAuthenticated:  graph.IsAuthenticated,
			RequiresSudoMode: graph.RequiresSudoMode,
		},
	}))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	if cfg.Environment == "development" {
		srv.Use(extension.Introspection{})
	}

	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	// Authentication middleware is already applied to all routes in the router
	r.Handle("/graphql", srv)

	r.Handle("/graphql/playground", playground.Handler("GraphQL Playground", "/graphql"))
}

func NewApp() *fx.App {
	config := config.SetupConfig()
	return fx.New(
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log}
		}),
		fx.Supply(config),
		fx.Provide(
			// router
			serverhttp.NewRouter,
			// database client
			db.NewDB,
			// s3 client
			s3client.NewS3ClientProvider,
			// logger
			logger.New,
			// GraphQL resolver
			resolver.NewResolver,
		),
		fx.Options(
			// Account domain repositories
			account.AccountDomainModule,
			// Auth domain repositories
			auth.AuthDomainModule,
		),
		fx.Invoke(
			AddGraphQLHandler,
			func(*chi.Mux) {},
		),
	)
}

func main() {
	NewApp().Run()
}
