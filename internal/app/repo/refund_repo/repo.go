// Package refund_repo contains database-related CRUD functionality.
package refund_repo

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/refund"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// queries.
var (
	//go:embed query/refund_create.sql
	refundCreateSQL string
	//go:embed query/refund_query.sql
	refundQuerySQL string
	//go:embed query/refund_query_by_id.sql
	refundQueryByIDSQL string
	//go:embed query/refund_query_by_payment_id.sql
	refundQueryByPaymentIDSQL string
)

// Store manages the set of APIs for refund database access.
type Store struct {
	log logger.Logger
	db  sqlx.ExtContext
}

// NewStore constructs the api for data access.
func NewStore(log logger.Logger, db *sqlx.DB) *Store {
	return &Store{
		log: log,
		db:  db,
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
		log: s.log,
		db:  ec,
	}

	return &store, nil
}

// Create inserts a new refund into the database.
func (s *Store) Create(ctx context.Context, ref refund.Refund) error {
	err := pgsql.NamedExecContext(ctx, s.log, s.db, refundCreateSQL, toRefundDB(ref))
	if err != nil {
		return fmt.Errorf("error refund create in db: %w", err)
	}

	return nil
}

// QueryByID gets the specified refund from the database.
func (s *Store) QueryByID(ctx context.Context, refundID uuid.UUID) (refund.Refund, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: refundID.String(),
	}

	var dbRef refundDB

	err := pgsql.NamedQueryStruct(ctx, s.log, s.db, refundQueryByIDSQL, data, &dbRef)
	if err != nil {
		if errors.Is(err, pgsql.ErrDBNotFound) {
			return refund.Refund{}, fmt.Errorf("db: %w", refund.ErrNotFound)
		}

		return refund.Refund{}, fmt.Errorf("db: %w", err)
	}

	return toRefundDomain(dbRef)
}

// QueryByPaymentID gets refunds by payment ID from the database.
func (s *Store) QueryByPaymentID(
	ctx context.Context,
	paymentID uuid.UUID,
) ([]refund.Refund, error) {
	data := struct {
		PaymentID string `db:"payment_id"`
	}{
		PaymentID: paymentID.String(),
	}

	var dbRefs []refundDB

	err := pgsql.NamedQuerySlice(ctx, s.log, s.db, refundQueryByPaymentIDSQL, data, &dbRefs)
	if err != nil {
		return nil, fmt.Errorf("error query refund by payment id in db: %w", err)
	}

	return toRefundsDomain(dbRefs)
}

// Query retrieves a list of existing refunds from the database.
func (s *Store) Query(
	ctx context.Context,
	filter refund.QueryFilter,
	orderBy order.By,
	cur cursor.Cursor,
) ([]refund.Refund, error) {
	data := map[string]any{
		"limit": cur.Limit() + 1,
	}

	buf := bytes.NewBufferString(refundQuerySQL)
	applyFilter(filter, data, buf)
	applyCursor(cur, orderBy, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" FETCH NEXT :limit ROWS ONLY")

	var dbRefs []refundDB

	err = pgsql.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbRefs)
	if err != nil {
		return nil, fmt.Errorf("error query refund in db: %w", err)
	}

	return toRefundsDomain(dbRefs)
}
