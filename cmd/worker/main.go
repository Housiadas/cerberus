package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/Housiadas/cerberus/internal/app/command"
	"github.com/Housiadas/cerberus/internal/config"
	ctxPck "github.com/Housiadas/cerberus/internal/utils/context"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/otel"
)

//nolint:gochecknoglobals
var build = "develop"

func main() {
	err := run()
	if err != nil {
		if !errors.Is(err, command.ErrHelp) {
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
	var log *logger.Service

	ctx := context.Background()
	traceIDFn := otel.GetTraceID(ctx)
	requestIDFn := ctxPck.GetRequestID(ctx)
	log = logger.New(os.Stdout, logger.LevelInfo, "Worker", traceIDFn, requestIDFn)

	// -------------------------------------------------------------------------
	// Start Tracing Support
	// -------------------------------------------------------------------------
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
		return fmt.Errorf("error starting tracing: %w", err)
	}
	defer teardown(ctx)

	tracer := traceProvider.Tracer(cfg.App.Name)

	// -------------------------------------------------------------------------
	// Initialize commands
	// -------------------------------------------------------------------------
	cmd := command.New(command.Config{
		DB:     cfg.DB,
		Kafka:  cfg.Kafka,
		Log:    log,
		Tracer: tracer,
		Version: config.Version{
			Desc:  "Worker",
			Build: build,
		},
	})

	return processCommands(os.Args, cmd)
}

// processCommands handles the execution of the commands specified on the command line.
func processCommands(args []string, cmd *command.Command) error {
	switch args[1] {
	case command.UserAdd:
		name := args[2]
		email := args[3]
		password := args[4]

		err := cmd.UserAdd(name, email, password)
		if err != nil {
			return fmt.Errorf("adding user: %w", err)
		}
	case command.OutboxRelay:
		err := cmd.OutboxRelay()
		if err != nil {
			return fmt.Errorf("outbox relay: %w", err)
		}

	default:
		fmt.Println("useradd:       add a new user to the database")
		fmt.Println("outbox-relay:  start the outbox relay process")
		fmt.Println("provide a command")

		return command.ErrHelp
	}

	return nil
}
