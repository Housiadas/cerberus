package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/Housiadas/cerberus/internal/app/command"
	"github.com/Housiadas/cerberus/internal/config"
	"github.com/Housiadas/cerberus/pkg/logger"
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
	c, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("parsing config: %w", err)
	}

	// -------------------------------------------------------------------------
	// Initialize Service
	// -------------------------------------------------------------------------
	traceIDFn := func(context.Context) string {
		return "00000000-0000-0000-0000-000000000000"
	}
	requestIDFn := func(context.Context) string {
		return "00000000-0000-0000-0000-000000000000"
	}
	log := logger.New(io.Discard, logger.LevelInfo, "CLI", traceIDFn, requestIDFn)

	// -------------------------------------------------------------------------
	// Initialize commands
	// -------------------------------------------------------------------------
	cmd := command.New(c, log, build, "CMD")

	return processCommands(os.Args, cmd)
}

// processCommands handles the execution of the commands specified on the command line.
func processCommands(args []string, cmd *command.Command) error {
	switch args[1] {
	case "useradd":
		name := args[2]
		email := args[3]

		password := args[4]

		err := cmd.UserAdd(name, email, password)
		if err != nil {
			return fmt.Errorf("adding user: %w", err)
		}

	default:
		fmt.Println("seed:       add data to the database")
		fmt.Println("useradd:    add a new user to the database")
		fmt.Println("provide a command to get more help.")

		return command.ErrHelp
	}

	return nil
}
