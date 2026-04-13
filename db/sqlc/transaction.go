package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync/atomic"

	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ctxKey is an unexported type used as a context key for transactions.
type ctxKey struct{}

// txState wraps a transaction with a done flag to prevent double-commit/rollback.
type txState struct {
	tx   *pgx.Tx
	done atomic.Bool
}

// Transactor provides transaction management by wrapping a database connection
// and delegating to RunInTx.
type Transactor struct {
	log  logger.Logger
	pool *pgxpool.Pool
}

// NewTransactor constructs a Transactor.
func NewTransactor(
	log logger.Logger,
	pool *pgxpool.Pool,
) *Transactor {
	return &Transactor{
		log:  log,
		pool: pool,
	}
}

// RunInTx begins a transaction, injects it into the context, and passes
// the enriched context to fn. If fn returns an error or the commit fails,
// the transaction is rolled back.
func (t *Transactor) RunInTx(
	ctx context.Context,
	fn func(ctx context.Context) error,
) error {
	// begin a transaction
	tx, err := t.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	st := &txState{tx: &tx}

	defer func() {
		// check atomic flag to prevent double-commit/rollback
		if st.done.Load() {
			return
		}

		rollbackErr := tx.Rollback(ctx)
		if rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
			t.log.Errorc(ctx, 5,
				"pgsql.RunInTx",
				"rollback error", rollbackErr,
			)
		}
	}()

	// inject transaction to context
	txCtx := context.WithValue(ctx, ctxKey{}, st)

	// run the callback function
	fnErr := fn(txCtx)
	if fnErr != nil {
		return fnErr
	}

	st.done.Store(true)

	// commit the transaction
	commitErr := tx.Commit(ctx)
	if commitErr != nil {
		return fmt.Errorf("commit transaction: %w", commitErr)
	}

	return nil
}
