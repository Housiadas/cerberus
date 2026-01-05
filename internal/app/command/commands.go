// Package commands contain the functionality for the set of commands
// currently supported by the CLI tooling.
package command

import (
	"errors"

	"github.com/Housiadas/cerberus/internal/config"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/pgsql"
)

// ErrHelp provides the context that help was given.
var ErrHelp = errors.New("help provided")

type Config struct {
	DB      config.DB
	Version config.Version
	Kafka   config.Kafka
}

type Command struct {
	DB      pgsql.Config
	Log     *logger.Service
	Version config.Version
	Kafka   config.Kafka
}

func New(
	cfg config.Config,
	log *logger.Service,
	build string,
	serviceName string,
) *Command {
	return &Command{
		DB: pgsql.Config{
			User:         cfg.DB.User,
			Password:     cfg.DB.Password,
			Host:         cfg.DB.Host,
			Name:         cfg.DB.Name,
			MaxIdleConns: cfg.DB.MaxIdleConns,
			MaxOpenConns: cfg.DB.MaxOpenConns,
			DisableTLS:   cfg.DB.DisableTLS,
		},
		Log: log,
		Version: config.Version{
			Build: build,
			Desc:  serviceName,
		},
	}
}
