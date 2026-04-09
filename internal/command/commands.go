// Package command contains the functionality
// for the set of commands currently supported by the Worker
package command

import (
	"github.com/Housiadas/cerberus/internal/config"
	"github.com/Housiadas/cerberus/pkg/email"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"go.opentelemetry.io/otel/trace"
)

const (
	UserAdd                = "user-add"
	OutboxRelay            = "outbox-relay"
	EmailNotificationRelay = "email-notification-relay"
)

type Config struct {
	DB      config.DB
	Kafka   config.Kafka
	Email   config.Email
	Log     logger.Logger
	Tracer  trace.Tracer
	Version config.Version
}

type Command struct {
	db          pgsql.Config
	log         logger.Logger
	tracer      trace.Tracer
	kafka       config.Kafka
	emailConfig email.Config
	version     config.Version
}

func New(cfg Config) *Command {
	return &Command{
		db: pgsql.Config{
			User:         cfg.DB.User,
			Password:     cfg.DB.Password,
			Host:         cfg.DB.Host,
			Name:         cfg.DB.Name,
			MaxIdleConns: cfg.DB.MinConns,
			MaxOpenConns: cfg.DB.MaxConns,
			DisableTLS:   cfg.DB.DisableTLS,
		},
		log:   cfg.Log,
		kafka: cfg.Kafka,
		emailConfig: email.Config{
			Host:     cfg.Email.Host,
			Port:     cfg.Email.Port,
			Username: cfg.Email.Username,
			Password: cfg.Email.Password,
			From:     cfg.Email.From,
			TLS:      cfg.Email.TLS,
		},
		version: config.Version{
			Build:       cfg.Version.Build,
			Description: cfg.Version.Description,
		},
	}
}
