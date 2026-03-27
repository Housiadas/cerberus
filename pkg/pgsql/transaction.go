package pgsql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// Beginner represents a value that can begin a transaction.
type Beginner interface {
	Begin() (CommitRollbacker, error)
}

// CommitRollbacker represents a value that can commit or rollback a transaction.
type CommitRollbacker interface {
	Commit() error
	Rollback() error
}

// DBBeginner implements the Beginner interface,
type DBBeginner struct {
	sqlxDB *sqlx.DB
}

// NewBeginner constructs a value that implements the beginner interface.
func NewBeginner(sqlxDB *sqlx.DB) *DBBeginner {
	return &DBBeginner{
		sqlxDB: sqlxDB,
	}
}

// Begin implements the Beginner interface and returns a concrete value that
// implements the CommitRollbacker interface.
func (db *DBBeginner) Begin() (CommitRollbacker, error) {
	tran, err := db.sqlxDB.Beginx()
	if err != nil {
		return nil, fmt.Errorf("begin transaction issue: %w", err)
	}

	return tran, nil
}

// GetExtContext is a helper function that extracts the sqlx value
// from the core transactor interface for transactional use.
func GetExtContext(tx CommitRollbacker) (sqlx.ExtContext, error) {
	ec, ok := tx.(sqlx.ExtContext)
	if !ok {
		return nil, fmt.Errorf("%w: got %T", ErrInvalidTransactorType, tx)
	}

	return ec, nil
}

// RunInTx begins a transaction, passes the handle to fn, then commits.
// If fn returns an error or the commit fails, the transaction is rolled back.
func RunInTx(
	ctx context.Context,
	log logger,
	beginner Beginner,
	fn func(CommitRollbacker) error,
) error {
	tran, err := beginner.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		rollbackErr := tran.Rollback()
		if rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
			log.Errorc(ctx, 5, "pgsql.RunInTx", "rollback error", rollbackErr)
		}
	}()

	fnErr := fn(tran)
	if fnErr != nil {
		return fnErr
	}

	commitErr := tran.Commit()
	if commitErr != nil {
		return fmt.Errorf("commit transaction: %w", commitErr)
	}

	return nil
}
