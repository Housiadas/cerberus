// Package command contains the functionality
// for the set of commands currently supported by the Worker
package command

import (
	"github.com/Housiadas/cerberus/internal/config"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"go.opentelemetry.io/otel/trace"
)

const (
	UserAdd     = "user-add"
	OutboxRelay = "outbox-relay"
)

type Config struct {
	DB      config.DB
	Kafka   config.Kafka
	Log     *logger.Service
	Tracer  trace.Tracer
	Version config.Version
}

type Command struct {
	db      pgsql.Config
	log     *logger.Service
	tracer  trace.Tracer
	kafka   config.Kafka
	version config.Version
}

func New(cfg Config) *Command {
	return &Command{
		db: pgsql.Config{
			User:         cfg.DB.User,
			Password:     cfg.DB.Password,
			Host:         cfg.DB.Host,
			Name:         cfg.DB.Name,
			MaxIdleConns: cfg.DB.MaxIdleConns,
			MaxOpenConns: cfg.DB.MaxOpenConns,
			DisableTLS:   cfg.DB.DisableTLS,
		},
		log:   cfg.Log,
		kafka: cfg.Kafka,
		version: config.Version{
			Build:       cfg.Version.Build,
			Description: cfg.Version.Description,
		},
	}
}
