// Package email_notification_outbox_repo contains database-related CRUD functionality for email notifications.
package email_notification_outbox_repo

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/internal/core/email_notification_outbox"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var (
	//go:embed query/email_outbox_create.sql
	emailOutboxCreateSQL string
	//go:embed query/email_outbox_query_unprocessed.sql
	emailOutboxQueryUnprocessedSQL string
	//go:embed query/email_outbox_mark_processed.sql
	emailOutboxMarkProcessedSQL string
	//go:embed query/email_outbox_increment_retry.sql
	emailOutboxIncrementRetrySQL string
)

type logger interface {
	Debug(ctx context.Context, msg string, args ...any)
	Debugc(ctx context.Context, caller int, msg string, args ...any)
	Info(ctx context.Context, msg string, args ...any)
	Infoc(ctx context.Context, caller int, msg string, args ...any)
	Warn(ctx context.Context, msg string, args ...any)
	Warnc(ctx context.Context, caller int, msg string, args ...any)
	Error(ctx context.Context, msg string, args ...any)
	Errorc(ctx context.Context, caller int, msg string, args ...any)
}

// Store manages the set of APIs for email notification outbox database access.
type Store struct {
	log    logger
	dbPool sqlx.ExtContext
}

// NewStore constructs the api for data access.
func NewStore(
	log logger,
	dbPool *sqlx.DB,
) *Store {
	return &Store{
		log:    log,
		dbPool: dbPool,
	}
}

// NewWithTx constructs a new Store value replacing the sqlx DB
// value with a sqlx DB value that is currently inside a transaction.
func (s *Store) NewWithTx(tx pgsql.CommitRollbacker) (email_notification_outbox.Storer, error) {
	ec, err := pgsql.GetExtContext(tx)
	if err != nil {
		return nil, fmt.Errorf("email_notification_outbox transaction init error: %w", err)
	}

	store := Store{
		log:    s.log,
		dbPool: ec,
	}

	return &store, nil
}

// Create inserts a new email notification outbox entry into the database.
func (s *Store) Create(
	ctx context.Context,
	e email_notification_outbox.EmailNotificationOutbox,
) error {
	err := pgsql.NamedExecContext(ctx, s.log, s.dbPool, emailOutboxCreateSQL, toEmailOutboxDB(e))
	if err != nil {
		return fmt.Errorf("named_exec_context: %w", err)
	}

	return nil
}

// QueryUnprocessed retrieves unprocessed email notification outbox entries.
func (s *Store) QueryUnprocessed(
	ctx context.Context,
	limit int,
	maxRetries int,
) ([]email_notification_outbox.EmailNotificationOutbox, error) {
	data := struct {
		Limit      int `db:"limit"`
		MaxRetries int `db:"max_retries"`
	}{
		Limit:      limit,
		MaxRetries: maxRetries,
	}

	var dbRows []emailOutboxDB

	err := pgsql.NamedQuerySlice(
		ctx,
		s.log,
		s.dbPool,
		emailOutboxQueryUnprocessedSQL,
		data,
		&dbRows,
	)
	if err != nil {
		return nil, fmt.Errorf("named_query_slice: %w", err)
	}

	result := make([]email_notification_outbox.EmailNotificationOutbox, len(dbRows))
	for i, row := range dbRows {
		result[i] = toEmailOutboxDomain(row)
	}

	return result, nil
}

// MarkProcessed updates email notification outbox entries as processed.
func (s *Store) MarkProcessed(
	ctx context.Context,
	ids []uuid.UUID,
	processedAt time.Time,
) error {
	data := struct {
		IDs         []uuid.UUID `db:"ids"`
		ProcessedAt time.Time   `db:"processed_at"`
	}{
		IDs:         ids,
		ProcessedAt: processedAt.UTC(),
	}

	named, args, err := sqlx.Named(emailOutboxMarkProcessedSQL, data)
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

// IncrementRetryCount increments the retry count for the given email outbox entry IDs.
func (s *Store) IncrementRetryCount(ctx context.Context, ids []uuid.UUID) error {
	data := struct {
		IDs []uuid.UUID `db:"ids"`
	}{
		IDs: ids,
	}

	named, args, err := sqlx.Named(emailOutboxIncrementRetrySQL, data)
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
