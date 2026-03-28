// Package refund_repo contains database-related CRUD functionality for refunds.
package refund_repo

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/refund"
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
	log logger
	db  *sqlx.DB
}

// NewStore constructs the api for data access.
func NewStore(
	log logger.Logger,
	db *sqlx.DB,
) *Store {
	return &Store{
		log: log,
		db:  db,
	}
}

// Create inserts a new refund into the database.
func (s *Store) Create(ctx context.Context, ref refund.Refund) error {
	err := pgsql.NamedExecContext(ctx, s.log, pgsql.Conn(ctx, s.db), refundCreateSQL, toRefundDB(ref))
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

	err := pgsql.NamedQuerySlice(ctx, s.log, pgsql.Conn(ctx, s.db), refundQueryByAccountIDSQL, data, &dbRefs)
	if err != nil {
		return nil, fmt.Errorf("named_query_slice: %w", err)
	}

	return toRefundsDomain(dbRefs), nil
}
