package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	command "github.com/Housiadas/cerberus/internal/command"
	"github.com/Housiadas/cerberus/internal/config"
	ctxPck "github.com/Housiadas/cerberus/internal/sdk/context"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/telemetry"
)

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
	log = logger.New(
		os.Stdout,
		logger.LevelInfo,
		"cerberus-worker",
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
	// Initialize commands
	// -------------------------------------------------------------------------
	cmd := command.New(command.Config{
		DB:            cfg.DB,
		Kafka:         cfg.Kafka,
		Email:         cfg.Email,
		Elasticsearch: cfg.Elasticsearch,
		Log:           log,
		Tracer:        tracer,
		Version: config.Version{
			Description: "Worker",
			Build:       build,
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

	case command.EmailNotificationRelay:
		err := cmd.EmailNotificationRelay()
		if err != nil {
			return fmt.Errorf("email notification relay: %w", err)
		}

	case command.ElasticSearchIndexer:
		err := cmd.Indexer()
		if err != nil {
			return fmt.Errorf("indexer: %w", err)
		}

	default:
		fmt.Println("useradd:                  add a new user to the database")
		fmt.Println("outbox-relay:             start the outbox relay process")
		fmt.Println("email-notification-relay: start the email notification relay process")
		fmt.Println("elasticsearch-indexer:    start the elasticsearch indexer process")
		fmt.Println("provide a command")

		return command.ErrHelp
	}

	return nil
}
