// Package cron contains the functionality
// for the set of cron jobs currently supported by the Cron runner.
package cron

import (
	"errors"
	"flag"
	"fmt"

	db "github.com/Housiadas/cerberus/db/sqlc"
	"github.com/Housiadas/cerberus/internal/config"
	"github.com/Housiadas/cerberus/pkg/logger"
	"go.opentelemetry.io/otel/trace"
)

const (
	OutboxCleanup = "outbox-cleanup"
)

// Runner defines the interface that all cron jobs must implement.
type Runner interface {
	Name() string
	Description() string
	SetupFlags(fs *flag.FlagSet)
	Run() error
}

// ErrHelp provides the context that help was given.
var ErrHelp = errors.New("help provided")

type Config struct {
	DB      config.DB
	Log     logger.Logger
	Tracer  trace.Tracer
	Version config.Version
}

type Cron struct {
	db      db.Config
	log     logger.Logger
	tracer  trace.Tracer
	version config.Version
}

func New(cfg Config) *Cron {
	return &Cron{
		db: db.Config{
			User:       cfg.DB.User,
			Password:   cfg.DB.Password,
			Host:       cfg.DB.Host,
			Name:       cfg.DB.Name,
			MinConns:   cfg.DB.MinConns,
			MaxConns:   cfg.DB.MaxConns,
			DisableTLS: cfg.DB.DisableTLS,
		},
		log:     cfg.Log,
		tracer:  cfg.Tracer,
		version: cfg.Version,
	}
}

// Registry returns the cron job registry.
func (c *Cron) Registry() map[string]Runner {
	runners := c.Runners()

	registry := make(map[string]Runner, len(runners))
	for _, r := range runners {
		registry[r.Name()] = r
	}

	return registry
}

func (c *Cron) PrintUsage() {
	fmt.Println("Usage: cron <job> [flags]")
	fmt.Println()
	fmt.Println("Jobs:")

	runners := c.Runners()
	for _, r := range runners {
		fmt.Printf("  %-26s %s\n", r.Name(), r.Description())
	}
}

// Runners returns all available cron job runners.
func (c *Cron) Runners() []Runner {
	return []Runner{
		NewOutboxCleanupRunner(c),
	}
}
