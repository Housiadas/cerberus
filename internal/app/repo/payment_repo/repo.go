// Package payment_repo contains database-related CRUD functionality.
package payment_repo

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/payment"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// queries.
var (
	//go:embed query/payment_create.sql
	paymentCreateSQL string
	//go:embed query/payment_update.sql
	paymentUpdateSQL string
	//go:embed query/payment_query.sql
	paymentQuerySQL string
	//go:embed query/payment_query_by_id.sql
	paymentQueryByIDSQL string
	//go:embed query/payment_query_by_invoice_id.sql
	paymentQueryByInvoiceIDSQL string
)

// Store manages the set of APIs for paymentDB database access.
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
func (s *Store) NewWithTx(tx pgsql.CommitRollbacker) (payment.Storer, error) {
	ec, err := pgsql.GetExtContext(tx)
	if err != nil {
		return nil, fmt.Errorf("payment transaction init error: %w", err)
	}

	store := Store{
		log: s.log,
		db:  ec,
	}

	return &store, nil
}

// Create inserts a new payment into the database.
func (s *Store) Create(ctx context.Context, pmt payment.Payment) error {
	err := pgsql.NamedExecContext(ctx, s.log, s.db, paymentCreateSQL, toPaymentDB(pmt))
	if err != nil {
		return fmt.Errorf("error payment create in db: %w", err)
	}

	return nil
}

// Update replaces a payment document in the database.
func (s *Store) Update(ctx context.Context, pmt payment.Payment) error {
	err := pgsql.NamedExecContext(ctx, s.log, s.db, paymentUpdateSQL, toPaymentDB(pmt))
	if err != nil {
		return fmt.Errorf("error payment update in db: %w", err)
	}

	return nil
}

// QueryByID gets the specified payment from the database.
func (s *Store) QueryByID(ctx context.Context, paymentID uuid.UUID) (payment.Payment, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: paymentID.String(),
	}

	var dbPmt paymentDB

	err := pgsql.NamedQueryStruct(ctx, s.log, s.db, paymentQueryByIDSQL, data, &dbPmt)
	if err != nil {
		if errors.Is(err, pgsql.ErrDBNotFound) {
			return payment.Payment{}, fmt.Errorf("db: %w", payment.ErrNotFound)
		}

		return payment.Payment{}, fmt.Errorf("db: %w", err)
	}

	return toDomain(dbPmt)
}

// QueryByInvoiceID retrieves payments for a specific invoice from the database.
func (s *Store) QueryByInvoiceID(ctx context.Context, invoiceID uuid.UUID) ([]payment.Payment, error) {
	data := struct {
		InvoiceID string `db:"invoice_id"`
	}{
		InvoiceID: invoiceID.String(),
	}

	var dbPmts []paymentDB

	err := pgsql.NamedQuerySlice(ctx, s.log, s.db, paymentQueryByInvoiceIDSQL, data, &dbPmts)
	if err != nil {
		return nil, fmt.Errorf("error query payment by invoice id in db: %w", err)
	}

	return toSliceDomain(dbPmts)
}

// Query retrieves a list of existing payments from the database.
func (s *Store) Query(
	ctx context.Context,
	filter payment.QueryFilter,
	orderBy order.By,
	cur cursor.Cursor,
) ([]payment.Payment, error) {
	data := map[string]any{
		"limit": cur.Limit() + 1,
	}

	buf := bytes.NewBufferString(paymentQuerySQL)
	applyFilter(filter, data, buf)
	applyCursor(cur, orderBy, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" FETCH NEXT :limit ROWS ONLY")

	var dbPmts []paymentDB

	err = pgsql.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbPmts)
	if err != nil {
		return nil, fmt.Errorf("error query payment in db: %w", err)
	}

	return toSliceDomain(dbPmts)
}
