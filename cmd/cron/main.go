package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/Housiadas/cerberus/internal/config"
	"github.com/Housiadas/cerberus/internal/cron"
	ctxPck "github.com/Housiadas/cerberus/internal/sdk/context"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/telemetry"
)

var build = "develop"

func main() {
	err := run()
	if err != nil {
		if !errors.Is(err, cron.ErrHelp) {
			fmt.Println("msg", err)
		}

		os.Exit(1)
	}
}

func run() error {
	// -------------------------------------------------------------------------
	// Initialize Configuration
	// -------------------------------------------------------------------------
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("parsing config: %w", err)
	}

	// -------------------------------------------------------------------------
	// Initialize Logger
	// -------------------------------------------------------------------------
	ctx := context.Background()
	log := logger.New(
		os.Stdout,
		logger.LevelInfo,
		"cerberus-cron",
		telemetry.GetTraceID,
		ctxPck.GetRequestID,
	)

	// -------------------------------------------------------------------------
	// Start Tracing Support
	// -------------------------------------------------------------------------
	tp, err := telemetry.New(ctx, telemetry.Config{
		ServiceName:  cfg.App.Name,
		OTLPEndpoint: cfg.Collector.Host,
		ExcludedRoutes: map[string]struct{}{
			"/liveness":  {},
			"/readiness": {},
		},
		TraceSampleRate: cfg.Collector.Probability,
		MetricInterval:  cfg.Collector.MetricInterval,
	})
	if err != nil {
		return fmt.Errorf("error starting tracing: %w", err)
	}
	defer tp.Shutdown(ctx) //nolint:errcheck

	tracer := tp.TracerProvider().Tracer(cfg.App.Name)

	// -------------------------------------------------------------------------
	// Initialize cron jobs
	// -------------------------------------------------------------------------
	c := cron.New(cron.Config{
		DB:     cfg.DB,
		Log:    log,
		Tracer: tracer,
		Version: config.Version{
			Description: "Cron",
			Build:       build,
		},
	})

	return processJobs(os.Args, c)
}

// processJobs handles the execution of the cron job specified on the command line.
func processJobs(args []string, c *cron.Cron) error {
	if len(args) < 2 {
		c.PrintUsage()

		return cron.ErrHelp
	}

	registry := c.Registry()

	runner, ok := registry[args[1]]
	if !ok {
		c.PrintUsage()

		return cron.ErrHelp
	}

	fs := flag.NewFlagSet(runner.Name(), flag.ContinueOnError)
	runner.SetupFlags(fs)

	err := fs.Parse(args[2:])
	if err != nil {
		return fmt.Errorf("parsing %s flags: %w", runner.Name(), err)
	}

	err = runner.Run()
	if err != nil {
		return fmt.Errorf("running %s: %w", runner.Name(), err)
	}

	return nil
}
