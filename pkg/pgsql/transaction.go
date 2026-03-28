package pgsql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync/atomic"

	"github.com/jmoiron/sqlx"
)

// ctxKey is an unexported type used as a context key for transactions.
type ctxKey struct{}

// txState wraps a transaction with a done flag to prevent double-commit/rollback.
type txState struct {
	tx   *sqlx.Tx
	done atomic.Bool
}

// Conn resolves a transaction from the context if present,
// otherwise falls back to db.
func Conn(ctx context.Context, db *sqlx.DB) sqlx.ExtContext {
	st, ok := ctx.Value(ctxKey{}).(*txState)
	if ok && !st.done.Load() {
		return st.tx
	}

	return db
}

// RunInTx begins a transaction, injects it into the context, and passes
// the enriched context to fn. If fn returns an error or the commit fails,
// the transaction is rolled back.
func RunInTx(
	ctx context.Context,
	log logger,
	db *sqlx.DB,
	fn func(ctx context.Context) error,
) error {
	sqlxTx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	st := &txState{tx: sqlxTx}

	defer func() {
		// check atomic flag to prevent double-commit/rollback
		if st.done.Load() {
			return
		}

		rollbackErr := sqlxTx.Rollback()
		if rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
			log.Errorc(ctx, 5, "pgsql.RunInTx", "rollback error", rollbackErr)
		}
	}()

	txCtx := context.WithValue(ctx, ctxKey{}, st)

	// run the callback function
	fnErr := fn(txCtx)
	if fnErr != nil {
		return fnErr
	}

	st.done.Store(true)

	// commit the transaction
	commitErr := sqlxTx.Commit()
	if commitErr != nil {
		return fmt.Errorf("commit transaction: %w", commitErr)
	}

	return nil
}
