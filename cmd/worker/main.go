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
	c, err := config.LoadConfig()
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
	// Initialize commands
	// -------------------------------------------------------------------------
	cmd := command.New(c, log, build, "Worker")

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
