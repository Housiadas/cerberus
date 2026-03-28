// Package invoice_repo contains database-related CRUD functionality for invoices.
package invoice_repo

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/invoice"
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

// Store manages the set of APIs for invoice database access.
type Store struct {
	log logger
	db  *sqlx.DB
}

// NewStore constructs the api for data access.
func NewStore(log logger, db *sqlx.DB) *Store {
	return &Store{
		log: log,
		db:  db,
	}
}

// Create inserts a new invoice into the database.
func (s *Store) Create(ctx context.Context, inv invoice.Invoice) error {
	err := pgsql.NamedExecContext(ctx, s.log, pgsql.Conn(ctx, s.db), invoiceCreateSQL, toInvoiceDB(inv))
	if err != nil {
		return fmt.Errorf("named_exec_context: %w", err)
	}

	return nil
}

// Update replaces an invoice in the database.
func (s *Store) Update(ctx context.Context, inv invoice.Invoice) error {
	err := pgsql.NamedExecContext(ctx, s.log, pgsql.Conn(ctx, s.db), invoiceUpdateSQL, toInvoiceDB(inv))
	if err != nil {
		return fmt.Errorf("named_exec_context: %w", err)
	}

	return nil
}

// QueryByAccountID retrieves invoices for an account.
func (s *Store) QueryByAccountID(
	ctx context.Context,
	accountID uuid.UUID,
) ([]invoice.Invoice, error) {
	data := struct {
		AccountID string `db:"account_id"`
	}{
		AccountID: accountID.String(),
	}

	var dbInvs []invoiceDB

	err := pgsql.NamedQuerySlice(ctx, s.log, pgsql.Conn(ctx, s.db), invoiceQueryByAccountIDSQL, data, &dbInvs)
	if err != nil {
		return nil, fmt.Errorf("named_query_slice: %w", err)
	}

	return toInvoicesDomain(dbInvs), nil
}

// QueryByStripeID retrieves an invoice by its Stripe ID.
func (s *Store) QueryByStripeID(ctx context.Context, stripeInvID string) (invoice.Invoice, error) {
	data := struct {
		StripeInvoiceID string `db:"stripe_invoice_id"`
	}{
		StripeInvoiceID: stripeInvID,
	}

	var dbInv invoiceDB

	err := pgsql.NamedQueryStruct(ctx, s.log, pgsql.Conn(ctx, s.db), invoiceQueryByStripeIDSQL, data, &dbInv)
	if err != nil {
		if errors.Is(err, pgsql.ErrDBNotFound) {
			return invoice.Invoice{}, fmt.Errorf("db: %w", invoice.ErrNotFound)
		}

		return invoice.Invoice{}, fmt.Errorf("db: %w", err)
	}

	return toInvoiceDomain(dbInv), nil
}
