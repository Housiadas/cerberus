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
	"time"

	db "github.com/Housiadas/cerberus/db/sqlc"
	"github.com/Housiadas/cerberus/internal/config"
	ctxPck "github.com/Housiadas/cerberus/internal/sdk/context"
	"github.com/Housiadas/cerberus/internal/sdk/elasticsearch"
	"github.com/Housiadas/cerberus/internal/sdk/host"
	"github.com/Housiadas/cerberus/internal/web/handler"
	"github.com/Housiadas/cerberus/pkg/debug"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/telemetry"
	"github.com/Housiadas/cerberus/pkg/vault"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"
)

var build = "develop"

func main() {
	// -------------------------------------------------------------------------
	// Initialize Logger
	// -------------------------------------------------------------------------
	var log *logger.Service

	ctx := context.Background()

	log = logger.New(
		os.Stdout,
		logger.LevelInfo,
		"cerberus-api",
		telemetry.GetTraceID,
		ctxPck.GetRequestID,
	)

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

	cfg.App.Version = config.Version{
		Build:       build,
		Description: "cerberus-api",
	}

	// -------------------------------------------------------------------------
	// App Starting
	// -------------------------------------------------------------------------
	log.Info(ctx, "startup", "GOMAXPROCS", runtime.GOMAXPROCS(0))

	log.Info(ctx, "starting application", "version", cfg.App.Version.Build)
	defer log.Info(ctx, "shutdown complete")

	log.BuildInfo(ctx)
	expvar.NewString("build").Set(cfg.App.Version.Build)

	// -------------------------------------------------------------------------
	// Initialize Telemetry
	// -------------------------------------------------------------------------
	tel, err := initTelemetry(ctx, cfg, log)
	if err != nil {
		return fmt.Errorf("initializing telemetry: %w", err)
	}

	defer tel.Shutdown(ctx) //nolint:errcheck

	// -------------------------------------------------------------------------
	// Initialize Database
	// -------------------------------------------------------------------------
	log.Info(ctx, "startup", "status", "initializing database", "host port", cfg.DB.Host)

	dbPool, err := db.Open(ctx, db.Config{
		User:       cfg.DB.User,
		Password:   cfg.DB.Password,
		Host:       cfg.DB.Host,
		Name:       cfg.DB.Name,
		Port:       cfg.DB.Port,
		MinConns:   cfg.DB.MinConns,
		MaxConns:   cfg.DB.MaxConns,
		DisableTLS: cfg.DB.DisableTLS,
	})
	if err != nil {
		return fmt.Errorf("connecting to db: %w", err)
	}
	defer dbPool.Close()

	// -------------------------------------------------------------------------
	// Initialize Redis
	// -------------------------------------------------------------------------
	log.Info(ctx, "startup", "status", "initializing redis", "host", cfg.Redis.Host)

	redisClient, redClose, err := initRedis(ctx, cfg)
	if err != nil {
		return fmt.Errorf("initializing redis: %w", err)
	}
	defer redClose()

	// -------------------------------------------------------------------------
	// Initialize Elasticsearch
	// -------------------------------------------------------------------------
	log.Info(ctx,
		"startup",
		"status", "initializing elasticsearch",
		"hosts", cfg.Elasticsearch.Hosts,
	)

	esClient, err := elasticsearch.New(log, elasticsearch.Config{
		Addresses: host.ParseHosts(cfg.Elasticsearch.Hosts),
		Username:  cfg.Elasticsearch.Username,
		Password:  cfg.Elasticsearch.Password,
	})
	if err != nil {
		log.Error(ctx,
			"elasticsearch unavailable, search disabled",
			"msg", err,
		)
	}

	// -------------------------------------------------------------------------
	// Initialize Vault Client
	// -------------------------------------------------------------------------
	jwtSecret, err := initVault(ctx, cfg)
	if err != nil {
		return fmt.Errorf("initializing vault: %w", err)
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
	h := handler.New(ctx, handler.Config{
		ServiceName:       cfg.App.Name,
		Build:             build,
		Cors:              cfg.Cors,
		DB:                dbPool,
		Redis:             redisClient,
		Log:               log,
		Tracer:            tel.TracerProvider().Tracer(cfg.App.Name),
		Meter:             tel.MeterProvider().Meter(cfg.App.Name),
		AccessTokenSecret: jwtSecret,
		FrontendURL:       cfg.App.FrontendURL,
		Stripe:            cfg.Stripe,
		ES:                esClient,
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

func initRedis(
	ctx context.Context,
	cfg config.Config,
) (*redis.Client, func(), error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Host,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	teardown := func() {
		client.Close()
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := client.Ping(pingCtx).Err()
	if err != nil {
		return nil, nil, fmt.Errorf("redis ping: %w", err)
	}

	return client, teardown, nil
}

func initVault(ctx context.Context, cfg config.Config) ([]byte, error) {
	vaultClient, err := vault.New(vault.Config{
		Address: cfg.Vault.Address,
		Token:   cfg.Vault.Token,
	})
	if err != nil {
		return nil, fmt.Errorf("creating vault client: %w", err)
	}

	jwtSecret, err := vaultClient.GetJWTSecret(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting jwt secret from vault: %w", err)
	}

	return jwtSecret, nil
}

func initTelemetry(
	ctx context.Context,
	cfg config.Config,
	log *logger.Service,
) (*telemetry.Provider, error) {
	log.Info(ctx, "startup", "status", "initializing telemetry")

	tel, err := telemetry.New(ctx, telemetry.Config{
		ServiceName:    cfg.App.Name,
		Environment:    cfg.App.Environment,
		Namespace:      cfg.App.Namespace,
		ServiceVersion: cfg.App.Version.Build,
		OTLPEndpoint:   cfg.Collector.Host,
		ExcludedRoutes: map[string]struct{}{
			"/liveness":  {},
			"/readiness": {},
		},
		TraceSampleRate: cfg.Collector.Probability,
		MetricInterval:  cfg.Collector.MetricInterval,
	})
	if err != nil {
		return nil, fmt.Errorf("error starting telemetry: %w", err)
	}

	return tel, nil
}
