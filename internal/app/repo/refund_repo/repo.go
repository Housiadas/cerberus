// Package refund_repo contains database related CRUD functionality for refunds.
package refund_repo

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/refund"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var (
	//go:embed query/refund_create.sql
	refundCreateSQL string
	//go:embed query/refund_query_by_account_id.sql
	refundQueryByAccountIDSQL string
)

// Store manages the set of APIs for refund database access.
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
func (s *Store) NewWithTx(tx pgsql.CommitRollbacker) (refund.Storer, error) {
	ec, err := pgsql.GetExtContext(tx)
	if err != nil {
		return nil, fmt.Errorf("refund transaction init error: %w", err)
	}

	store := Store{
		log:    s.log,
		dbPool: ec,
	}

	return &store, nil
}

// Create inserts a new refund into the database.
func (s *Store) Create(ctx context.Context, ref refund.Refund) error {
	err := pgsql.NamedExecContext(ctx, s.log, s.dbPool, refundCreateSQL, toRefundDB(ref))
	if err != nil {
		return fmt.Errorf("named_exec_context: %w", err)
	}

	return nil
}

// QueryByAccountID retrieves refunds for an account.
func (s *Store) QueryByAccountID(
	ctx context.Context,
	accountID uuid.UUID,
) ([]refund.Refund, error) {
	data := struct {
		AccountID string `db:"account_id"`
	}{
		AccountID: accountID.String(),
	}

	var dbRefs []refundDB

	err := pgsql.NamedQuerySlice(ctx, s.log, s.dbPool, refundQueryByAccountIDSQL, data, &dbRefs)
	if err != nil {
		return nil, fmt.Errorf("named_query_slice: %w", err)
	}

	return toRefundsDomain(dbRefs), nil
}
