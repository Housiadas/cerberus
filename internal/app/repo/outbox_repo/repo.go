// Package outbox_repo contains database related CRUD functionality for outbox.
package outbox_repo

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/outbox"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var (
	//go:embed query/outbox_create.sql
	outboxCreateSQL string
	//go:embed query/outbox_query_unprocessed.sql
	outboxQueryUnprocessedSQL string
	//go:embed query/outbox_mark_processed.sql
	outboxMarkProcessedSQL string
	//go:embed query/outbox_increment_retry.sql
	outboxIncrementRetrySQL string
)

// Store manages the set of APIs for outbox database access.
type Store struct {
	log    logger.Logger
	dbPool sqlx.ExtContext
}

// NewStore constructs the api for data access.
func NewStore(log logger.Logger, dbPool *sqlx.DB) *Store {
	return &Store{
		log:    log,
		dbPool: dbPool,
	}
}

// NewWithTx constructs a new Store value replacing the sqlx DB
// value with a sqlx DB value that is currently inside a transaction.
func (s *Store) NewWithTx(tx pgsql.CommitRollbacker) (outbox.Storer, error) {
	ec, err := pgsql.GetExtContext(tx)
	if err != nil {
		return nil, fmt.Errorf("outbox transaction init error: %w", err)
	}

	store := Store{
		log:    s.log,
		dbPool: ec,
	}

	return &store, nil
}

// Create inserts a new outbox entry into the database.
func (s *Store) Create(ctx context.Context, o outbox.Outbox) error {
	err := pgsql.NamedExecContext(ctx, s.log, s.dbPool, outboxCreateSQL, toOutboxDB(o))
	if err != nil {
		return fmt.Errorf("named_exec_context: %w", err)
	}

	return nil
}

// QueryUnprocessed retrieves unprocessed outbox entries below the max retry threshold.
func (s *Store) QueryUnprocessed(
	ctx context.Context,
	limit int,
	maxRetries int,
) ([]outbox.Outbox, error) {
	data := struct {
		Limit      int `db:"limit"`
		MaxRetries int `db:"max_retries"`
	}{
		Limit:      limit,
		MaxRetries: maxRetries,
	}

	var dbRows []outboxDB

	err := pgsql.NamedQuerySlice(ctx, s.log, s.dbPool, outboxQueryUnprocessedSQL, data, &dbRows)
	if err != nil {
		return nil, fmt.Errorf("named_query_slice: %w", err)
	}

	result := make([]outbox.Outbox, len(dbRows))
	for i, row := range dbRows {
		result[i] = toOutboxDomain(row)
	}

	return result, nil
}

// MarkProcessed updates outbox entries as processed.
func (s *Store) MarkProcessed(ctx context.Context, ids []uuid.UUID, processedAt time.Time) error {
	data := struct {
		IDs         []uuid.UUID `db:"ids"`
		ProcessedAt time.Time   `db:"processed_at"`
	}{
		IDs:         ids,
		ProcessedAt: processedAt.UTC(),
	}

	named, args, err := sqlx.Named(outboxMarkProcessedSQL, data)
	if err != nil {
		return fmt.Errorf("sqlx named error: %w", err)
	}

	query, args, err := sqlx.In(named, args...)
	if err != nil {
		return fmt.Errorf("sqlx in error: %w", err)
	}

	query = s.dbPool.Rebind(query)

	_, err = s.dbPool.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("exec_context: %w", err)
	}

	return nil
}

// IncrementRetryCount increments the retry count for the given outbox entry IDs.
func (s *Store) IncrementRetryCount(ctx context.Context, ids []uuid.UUID) error {
	data := struct {
		IDs []uuid.UUID `db:"ids"`
	}{
		IDs: ids,
	}

	named, args, err := sqlx.Named(outboxIncrementRetrySQL, data)
	if err != nil {
		return fmt.Errorf("sqlx named error: %w", err)
	}

	query, args, err := sqlx.In(named, args...)
	if err != nil {
		return fmt.Errorf("sqlx in error: %w", err)
	}

	query = s.dbPool.Rebind(query)

	_, err = s.dbPool.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("exec_context: %w", err)
	}

	return nil
}
