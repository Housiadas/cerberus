package main

import (
	"context"
	"expvar"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	_ "github.com/Housiadas/cerberus/docs"
	"github.com/Housiadas/cerberus/internal/app/handler"
	ctxPck "github.com/Housiadas/cerberus/internal/common/context"
	"github.com/Housiadas/cerberus/internal/config"
	"github.com/Housiadas/cerberus/pkg/debug"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/otel"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/Housiadas/cerberus/pkg/vault"
)

//nolint:gochecknoglobals
var build = "develop"

// @title						Cerberus
// @description				This is a monitoring system.
//
// @contact.name				API Support
// @contact.url				http://www.swagger.io/support
// @contact.email				support@swagger.io
//
// @license.name				Apache 2.0
// @license.url				http://www.apache.org/licenses/LICENSE-2.0.html
//
// @query.collection.format	multi
//
// @externalDocs.description	OpenAPI
//
// @externalDocs.url			https://swagger.io/resources/open-api/
// @host						localhost:4000.
func main() {
	// -------------------------------------------------------------------------
	// Initialize Logger
	// -------------------------------------------------------------------------
	var log *logger.Service

	ctx := context.Background()

	traceIDFn := otel.GetTraceID(ctx)
	requestIDFn := ctxPck.GetRequestID(ctx)
	log = logger.New(os.Stdout, logger.LevelInfo, "Rest api", traceIDFn, requestIDFn)

	// -------------------------------------------------------------------------
	// Run the application
	// -------------------------------------------------------------------------
	err := run(ctx, log)
	if err != nil {
		log.Error(ctx, "error during rest server startup", "msg", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, log *logger.Service) error {
	// -------------------------------------------------------------------------
	// Initialize Configuration
	// -------------------------------------------------------------------------
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Error(ctx, "error during config initialization", "msg", err)
		os.Exit(1)
	}

	cfg.Version = config.Version{
		Build: build,
		Desc:  "API",
	}

	// -------------------------------------------------------------------------
	// App Starting
	// -------------------------------------------------------------------------
	log.Info(ctx, "startup", "GOMAXPROCS", runtime.GOMAXPROCS(0))

	log.Info(ctx, "starting application", "version", cfg.Version.Build)
	defer log.Info(ctx, "shutdown complete")

	log.BuildInfo(ctx)
	expvar.NewString("build").Set(cfg.Version.Build)

	// -------------------------------------------------------------------------
	// Initialize Database
	// -------------------------------------------------------------------------
	log.Info(ctx, "startup", "status", "initializing database", "host port", cfg.DB.Host)

	db, err := pgsql.Open(pgsql.Config{
		User:         cfg.DB.User,
		Password:     cfg.DB.Password,
		Host:         cfg.DB.Host,
		Name:         cfg.DB.Name,
		MaxIdleConns: cfg.DB.MaxIdleConns,
		MaxOpenConns: cfg.DB.MaxOpenConns,
		DisableTLS:   cfg.DB.DisableTLS,
	})
	if err != nil {
		return fmt.Errorf("connecting to db: %w", err)
	}

	defer db.Close()

	// -------------------------------------------------------------------------
	// Start Tracing Support
	// -------------------------------------------------------------------------
	log.Info(ctx, "startup", "status", "initializing tracing support")

	traceProvider, teardown, err := otel.InitTracing(ctx, otel.Config{
		ServiceName: cfg.App.Name,
		Host:        cfg.Tempo.Host,
		ExcludedRoutes: map[string]struct{}{
			"/liveness":  {},
			"/readiness": {},
		},
		Probability: cfg.Tempo.Probability,
	})
	if err != nil {
		return fmt.Errorf("starting tracing: %w", err)
	}

	defer teardown(ctx)

	tracer := traceProvider.Tracer(cfg.App.Name)

	// -------------------------------------------------------------------------
	// Initialize Vault Client
	// -------------------------------------------------------------------------
	log.Info(ctx, "startup", "status", "initializing vault client", "address", cfg.Vault.Address)

	vaultClient, err := vault.New(vault.Config{
		Address: cfg.Vault.Address,
		Token:   cfg.Vault.Token,
	})
	if err != nil {
		return fmt.Errorf("creating vault client: %w", err)
	}

	jwtSecret, err := vaultClient.GetJWTSecret(ctx)
	if err != nil {
		return fmt.Errorf("getting jwt secret from vault: %w", err)
	}

	log.Info(ctx, "startup", "status", "jwt secret loaded from vault")

	// -------------------------------------------------------------------------
	// Start Debug Rest Server
	// -------------------------------------------------------------------------
	go func() {
		log.Info(ctx, "startup", "status", "Debug server starting", "host", cfg.Rest.Debug)

		debugSrv := http.Server{
			Addr:         cfg.Rest.Debug,
			Handler:      debug.Mux(),
			ReadTimeout:  cfg.Rest.ReadTimeout,
			WriteTimeout: cfg.Rest.WriteTimeout,
			IdleTimeout:  cfg.Rest.IdleTimeout,
		}

		err := debugSrv.ListenAndServe()
		if err != nil {
			log.Error(ctx, "shutdown",
				"status", "debug router closed",
				"host", cfg.Rest.Debug,
				"msg", err,
			)
		}
	}()

	// -------------------------------------------------------------------------
	// Start API Rest Server
	// -------------------------------------------------------------------------
	log.Info(ctx, "startup", "status", "Rest server starting")

	// Initialize handler
	h := handler.New(handler.Config{
		ServiceName:       cfg.App.Name,
		Build:             build,
		Cors:              cfg.Cors,
		DB:                db,
		Log:               log,
		Tracer:            tracer,
		AccessTokenSecret: jwtSecret,
	})

	api := http.Server{
		Addr:         cfg.Rest.API,
		Handler:      h.Routes(),
		ReadTimeout:  cfg.Rest.ReadTimeout,
		WriteTimeout: cfg.Rest.WriteTimeout,
		IdleTimeout:  cfg.Rest.IdleTimeout,
		ErrorLog:     logger.NewStdLogger(log, logger.LevelError),
	}

	serverErrors := make(chan error, 1)
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Info(ctx, "startup", "status", "Rest server started", "host", api.Addr)

		serverErrors <- api.ListenAndServe()
	}()

	// Shutdown
	select {
	case err := <-serverErrors:
		return fmt.Errorf("rest server error: %w", err)

	case sig := <-shutdown:
		log.Info(ctx, "shutdown", "status", "shutdown started", "signal", sig)
		defer log.Info(ctx, "shutdown", "status", "shutdown complete", "signal", sig)

		ctx, cancel := context.WithTimeout(ctx, cfg.Rest.ShutdownTimeout)
		defer cancel()

		err := api.Shutdown(ctx)
		if err != nil {
			api.Close()

			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
	}

	return nil
}
