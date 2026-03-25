// Package invoice_repo contains database related CRUD functionality for invoices.
package invoice_repo

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	invoice2 "github.com/Housiadas/cerberus/internal/core/invoice"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var (
	//go:embed query/invoice_create.sql
	invoiceCreateSQL string
	//go:embed query/invoice_update.sql
	invoiceUpdateSQL string
	//go:embed query/invoice_query_by_account_id.sql
	invoiceQueryByAccountIDSQL string
	//go:embed query/invoice_query_by_stripe_id.sql
	invoiceQueryByStripeIDSQL string
)

// Store manages the set of APIs for invoice database access.
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
func (s *Store) NewWithTx(tx pgsql.CommitRollbacker) (invoice2.Storer, error) {
	ec, err := pgsql.GetExtContext(tx)
	if err != nil {
		return nil, fmt.Errorf("invoice transaction init error: %w", err)
	}

	store := Store{
		log:    s.log,
		dbPool: ec,
	}

	return &store, nil
}

// Create inserts a new invoice into the database.
func (s *Store) Create(ctx context.Context, inv invoice2.Invoice) error {
	err := pgsql.NamedExecContext(ctx, s.log, s.dbPool, invoiceCreateSQL, toInvoiceDB(inv))
	if err != nil {
		return fmt.Errorf("named_exec_context: %w", err)
	}

	return nil
}

// Update replaces an invoice in the database.
func (s *Store) Update(ctx context.Context, inv invoice2.Invoice) error {
	err := pgsql.NamedExecContext(ctx, s.log, s.dbPool, invoiceUpdateSQL, toInvoiceDB(inv))
	if err != nil {
		return fmt.Errorf("named_exec_context: %w", err)
	}

	return nil
}

// QueryByAccountID retrieves invoices for an account.
func (s *Store) QueryByAccountID(
	ctx context.Context,
	accountID uuid.UUID,
) ([]invoice2.Invoice, error) {
	data := struct {
		AccountID string `db:"account_id"`
	}{
		AccountID: accountID.String(),
	}

	var dbInvs []invoiceDB

	err := pgsql.NamedQuerySlice(ctx, s.log, s.dbPool, invoiceQueryByAccountIDSQL, data, &dbInvs)
	if err != nil {
		return nil, fmt.Errorf("named_query_slice: %w", err)
	}

	return toInvoicesDomain(dbInvs), nil
}

// QueryByStripeID retrieves an invoice by its Stripe ID.
func (s *Store) QueryByStripeID(ctx context.Context, stripeInvID string) (invoice2.Invoice, error) {
	data := struct {
		StripeInvoiceID string `db:"stripe_invoice_id"`
	}{
		StripeInvoiceID: stripeInvID,
	}

	var dbInv invoiceDB

	err := pgsql.NamedQueryStruct(ctx, s.log, s.dbPool, invoiceQueryByStripeIDSQL, data, &dbInv)
	if err != nil {
		if errors.Is(err, pgsql.ErrDBNotFound) {
			return invoice2.Invoice{}, fmt.Errorf("db: %w", invoice2.ErrNotFound)
		}

		return invoice2.Invoice{}, fmt.Errorf("db: %w", err)
	}

	return toInvoiceDomain(dbInv), nil
}
