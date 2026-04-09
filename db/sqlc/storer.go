package db

import (
	"context"

	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/order"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Builder is the squirrel statement builder configured for Postgres ($N
// placeholders). Hand-written dynamic queries in this package use it.
var Builder = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

// Storer combines the sqlc-generated static Querier with hand-written dynamic
// queries (built with squirrel) so callers see a single interface for all
// account-store operations.
type Storer interface {
	Querier

	QueryAccounts(
		ctx context.Context,
		filter AccountQueryFilter,
		orderBy order.By,
		cur cursor.Cursor,
	) ([]Account, error)
}

// Store is the concrete Storer. It embeds the sqlc-generated *Queries (which
// supplies every static CRUD method) and holds the active DBTX so dynamic
// queries hit the same connection — pool by default, pgx.Tx after WithTx.
type Store struct {
	*Queries
	pool *pgxpool.Pool
	db   DBTX
}

// NewStorer creates a Store backed by the given pgx connection pool.
func NewStorer(pool *pgxpool.Pool) *Store {
	return &Store{
		Queries: New(pool),
		pool:    pool,
		db:      pool,
	}
}

// WithTx returns a Store whose static and dynamic queries both run on tx.
func (s *Store) WithTx(tx pgx.Tx) *Store {
	return &Store{
		Queries: s.Queries.WithTx(tx),
		pool:    s.pool,
		db:      tx,
	}
}
