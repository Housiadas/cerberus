// Package payment_repo contains database related CRUD functionality for payments.
package payment_repo

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	payment2 "github.com/Housiadas/cerberus/internal/core/payment"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var (
	//go:embed query/payment_create.sql
	paymentCreateSQL string
	//go:embed query/payment_update.sql
	paymentUpdateSQL string
	//go:embed query/payment_query_by_account_id.sql
	paymentQueryByAccountIDSQL string
	//go:embed query/payment_query_by_stripe_id.sql
	paymentQueryByStripeIDSQL string
)

// Store manages the set of APIs for payment database access.
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
func (s *Store) NewWithTx(tx pgsql.CommitRollbacker) (payment2.Storer, error) {
	ec, err := pgsql.GetExtContext(tx)
	if err != nil {
		return nil, fmt.Errorf("payment transaction init error: %w", err)
	}

	store := Store{
		log:    s.log,
		dbPool: ec,
	}

	return &store, nil
}

// Create inserts a new payment into the database.
func (s *Store) Create(ctx context.Context, pmt payment2.Payment) error {
	err := pgsql.NamedExecContext(ctx, s.log, s.dbPool, paymentCreateSQL, toPaymentDB(pmt))
	if err != nil {
		return fmt.Errorf("named_exec_context: %w", err)
	}

	return nil
}

// Update replaces a payment in the database.
func (s *Store) Update(ctx context.Context, pmt payment2.Payment) error {
	err := pgsql.NamedExecContext(ctx, s.log, s.dbPool, paymentUpdateSQL, toPaymentDB(pmt))
	if err != nil {
		return fmt.Errorf("named_exec_context: %w", err)
	}

	return nil
}

// QueryByAccountID retrieves payments for an account.
func (s *Store) QueryByAccountID(
	ctx context.Context,
	accountID uuid.UUID,
) ([]payment2.Payment, error) {
	data := struct {
		AccountID string `db:"account_id"`
	}{
		AccountID: accountID.String(),
	}

	var dbPmts []paymentDB

	err := pgsql.NamedQuerySlice(ctx, s.log, s.dbPool, paymentQueryByAccountIDSQL, data, &dbPmts)
	if err != nil {
		return nil, fmt.Errorf("named_query_slice: %w", err)
	}

	return toPaymentsDomain(dbPmts), nil
}

// QueryByStripeID retrieves a payment by its Stripe ID.
func (s *Store) QueryByStripeID(
	ctx context.Context,
	stripePaymentID string,
) (payment2.Payment, error) {
	data := struct {
		StripePaymentID string `db:"stripe_payment_id"`
	}{
		StripePaymentID: stripePaymentID,
	}

	var dbPmt paymentDB

	err := pgsql.NamedQueryStruct(ctx, s.log, s.dbPool, paymentQueryByStripeIDSQL, data, &dbPmt)
	if err != nil {
		if errors.Is(err, pgsql.ErrDBNotFound) {
			return payment2.Payment{}, fmt.Errorf("db: %w", payment2.ErrNotFound)
		}

		return payment2.Payment{}, fmt.Errorf("db: %w", err)
	}

	return toPaymentDomain(dbPmt), nil
}
