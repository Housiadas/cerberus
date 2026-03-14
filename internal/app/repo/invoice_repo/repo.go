// Package invoice_repo contains database-related CRUD functionality.
package invoice_repo

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/invoice"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// queries.
var (
	//go:embed query/invoice_create.sql
	invoiceCreateSQL string
	//go:embed query/invoice_update.sql
	invoiceUpdateSQL string
	//go:embed query/invoice_query.sql
	invoiceQuerySQL string
	//go:embed query/invoice_query_by_id.sql
	invoiceQueryByIDSQL string
	//go:embed query/invoice_query_by_account_id.sql
	invoiceQueryByAccountIDSQL string
)

// Store manages the set of APIs for invoiceDB database access.
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
func (s *Store) NewWithTx(tx pgsql.CommitRollbacker) (invoice.Storer, error) {
	ec, err := pgsql.GetExtContext(tx)
	if err != nil {
		return nil, fmt.Errorf("invoice transaction init error: %w", err)
	}

	store := Store{
		log: s.log,
		db:  ec,
	}

	return &store, nil
}

// Create inserts a new invoice into the database.
func (s *Store) Create(ctx context.Context, inv invoice.Invoice) error {
	err := pgsql.NamedExecContext(ctx, s.log, s.db, invoiceCreateSQL, toInvoiceDB(inv))
	if err != nil {
		return fmt.Errorf("error invoice create in db: %w", err)
	}

	return nil
}

// Update replaces an invoice document in the database.
func (s *Store) Update(ctx context.Context, inv invoice.Invoice) error {
	err := pgsql.NamedExecContext(ctx, s.log, s.db, invoiceUpdateSQL, toInvoiceDB(inv))
	if err != nil {
		return fmt.Errorf("error invoice update in db: %w", err)
	}

	return nil
}

// QueryByID gets the specified invoice from the database.
func (s *Store) QueryByID(ctx context.Context, invoiceID uuid.UUID) (invoice.Invoice, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: invoiceID.String(),
	}

	var dbInv invoiceDB

	err := pgsql.NamedQueryStruct(ctx, s.log, s.db, invoiceQueryByIDSQL, data, &dbInv)
	if err != nil {
		if errors.Is(err, pgsql.ErrDBNotFound) {
			return invoice.Invoice{}, fmt.Errorf("db: %w", invoice.ErrNotFound)
		}

		return invoice.Invoice{}, fmt.Errorf("db: %w", err)
	}

	return toDomain(dbInv)
}

// QueryByAccountID retrieves invoices for a specific account from the database.
func (s *Store) QueryByAccountID(ctx context.Context, accountID uuid.UUID) ([]invoice.Invoice, error) {
	data := struct {
		AccountID string `db:"account_id"`
	}{
		AccountID: accountID.String(),
	}

	var dbInvs []invoiceDB

	err := pgsql.NamedQuerySlice(ctx, s.log, s.db, invoiceQueryByAccountIDSQL, data, &dbInvs)
	if err != nil {
		return nil, fmt.Errorf("error query invoice by account id in db: %w", err)
	}

	return toSliceDomain(dbInvs)
}

// Query retrieves a list of existing invoices from the database.
func (s *Store) Query(
	ctx context.Context,
	filter invoice.QueryFilter,
	orderBy order.By,
	cur cursor.Cursor,
) ([]invoice.Invoice, error) {
	data := map[string]any{
		"limit": cur.Limit() + 1,
	}

	buf := bytes.NewBufferString(invoiceQuerySQL)
	applyFilter(filter, data, buf)
	applyCursor(cur, orderBy, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" FETCH NEXT :limit ROWS ONLY")

	var dbInvs []invoiceDB

	err = pgsql.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbInvs)
	if err != nil {
		return nil, fmt.Errorf("error query invoice in db: %w", err)
	}

	return toSliceDomain(dbInvs)
}
