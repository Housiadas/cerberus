package db

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Builder is the squirrel statement builder configured for Postgres ($N
// placeholders). Handwritten dynamic queries in this package use it.
var Builder = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

// Store embeds the sqlc-generated *Queries (which
// supplies every static CRUD method) and holds the active DBTX, so dynamic
// queries hit the same connection — pool by default, pgx.Tx after WithTx.
type Store struct {
	*Queries
	pool *pgxpool.Pool
}

// NewStorer creates a Store backed by the given pgx connection pool.
func NewStorer(pool *pgxpool.Pool) *Store {
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
