package db

import (
	"context"
	"crypto/tls"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Builder is the squirrel statement builder configured for Postgres ($N
// placeholders). Handwritten dynamic queries in this package use it.
var Builder = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

// Config is the required properties to use the database.
type Config struct {
	User       string
	Password   string
	Host       string
	Name       string
	Port       uint16
	MinConns   int
	MaxConns   int
	DisableTLS bool
}

// Store embeds the sqlc-generated *Queries (which
// supplies every static CRUD method) and holds the active DBTX, so dynamic
// queries hit the same connection — pool by default, pgx.Tx after WithTx.
type Store struct {
	*Queries
	pool *pgxpool.Pool
}

// NewStore creates a Store backed by the given pgx connection pool.
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{
		Queries: New(pool),
		pool:    pool,
	}
}

// WithTx returns a Store whose static and dynamic queries both run on tx.
func (s *Store) WithTx(tx pgx.Tx) *Store {
	return &Store{
		Queries: s.Queries.WithTx(tx),
		pool:    s.pool,
	}
}

// Open a database connection pool based on the configuration. The pool is
// instrumented with an OpenTelemetry tracer, so every query is recorded as a span.
func Open(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	poolCfg := &pgxpool.Config{
		ConnConfig: &pgx.ConnConfig{
			Config: pgconn.Config{
				Host:     cfg.Host,
				Port:     cfg.Port,
				Database: cfg.Name,
				User:     cfg.User,
				Password: cfg.Password,
			},
			Tracer: otelpgx.NewTracer(),
		},
		MaxConns: int32(cfg.MaxConns),
		MinConns: int32(cfg.MinConns),
	}

	if cfg.DisableTLS {
		poolCfg.ConnConfig.TLSConfig = nil
	} else {
		poolCfg.ConnConfig.TLSConfig = &tls.Config{
			ServerName: cfg.Host,
			MinVersion: tls.VersionTLS12,
		}
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("pgx driver open error: %w", err)
	}

	return pool, nil
}

// Status pings the database and is intended for readiness probes. It returns
// nil when the pool can serve queries.
func (s *Store) Status(ctx context.Context) error {
	err := s.pool.Ping(ctx)
	if err != nil {
		return fmt.Errorf("ping database: %w", err)
	}

	return nil
}
