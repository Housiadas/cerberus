// Package billing_address_repo contains database-related CRUD functionality.
package billing_address_repo

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/billing_address"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// queries.
var (
	//go:embed query/billing_address_create.sql
	billingAddressCreateSQL string
	//go:embed query/billing_address_update.sql
	billingAddressUpdateSQL string
	//go:embed query/billing_address_query_by_id.sql
	billingAddressQueryByIDSQL string
	//go:embed query/billing_address_query_by_account_id.sql
	billingAddressQueryByAccountIDSQL string
)

// Store manages the set of APIs for billing_address database access.
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
func (s *Store) NewWithTx(tx pgsql.CommitRollbacker) (billing_address.Storer, error) {
	ec, err := pgsql.GetExtContext(tx)
	if err != nil {
		return nil, fmt.Errorf("billing address transaction init error: %w", err)
	}

	store := Store{
		log: s.log,
		db:  ec,
	}

	return &store, nil
}

// Create inserts a new billing address into the database.
func (s *Store) Create(ctx context.Context, ba billing_address.BillingAddress) error {
	err := pgsql.NamedExecContext(ctx, s.log, s.db, billingAddressCreateSQL, toBillingAddressDB(ba))
	if err != nil {
		return fmt.Errorf("error billing address create in db: %w", err)
	}

	return nil
}

// Update replaces a billing address document in the database.
func (s *Store) Update(ctx context.Context, ba billing_address.BillingAddress) error {
	err := pgsql.NamedExecContext(ctx, s.log, s.db, billingAddressUpdateSQL, toBillingAddressDB(ba))
	if err != nil {
		return fmt.Errorf("error billing address update in db: %w", err)
	}

	return nil
}

// QueryByID gets the specified billing address from the database.
func (s *Store) QueryByID(ctx context.Context, billingAddressID uuid.UUID) (billing_address.BillingAddress, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: billingAddressID.String(),
	}

	var dbBA billingAddressDB

	err := pgsql.NamedQueryStruct(ctx, s.log, s.db, billingAddressQueryByIDSQL, data, &dbBA)
	if err != nil {
		if errors.Is(err, pgsql.ErrDBNotFound) {
			return billing_address.BillingAddress{}, fmt.Errorf("db: %w", billing_address.ErrNotFound)
		}

		return billing_address.BillingAddress{}, fmt.Errorf("db: %w", err)
	}

	return toBillingAddressDomain(dbBA)
}

// QueryByAccountID gets the billing address by account ID from the database.
func (s *Store) QueryByAccountID(ctx context.Context, accountID uuid.UUID) (billing_address.BillingAddress, error) {
	data := struct {
		AccountID string `db:"account_id"`
	}{
		AccountID: accountID.String(),
	}

	var dbBA billingAddressDB

	err := pgsql.NamedQueryStruct(ctx, s.log, s.db, billingAddressQueryByAccountIDSQL, data, &dbBA)
	if err != nil {
		if errors.Is(err, pgsql.ErrDBNotFound) {
			return billing_address.BillingAddress{}, fmt.Errorf("db: %w", billing_address.ErrNotFound)
		}

		return billing_address.BillingAddress{}, fmt.Errorf("db: %w", err)
	}

	return toBillingAddressDomain(dbBA)
}
