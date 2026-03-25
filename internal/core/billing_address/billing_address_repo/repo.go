// Package billing_address_repo contains database related CRUD functionality for billing addresses.
package billing_address_repo

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	billing_address2 "github.com/Housiadas/cerberus/internal/core/billing_address"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var (
	//go:embed query/billing_address_create.sql
	billingAddressCreateSQL string
	//go:embed query/billing_address_update.sql
	billingAddressUpdateSQL string
	//go:embed query/billing_address_delete.sql
	billingAddressDeleteSQL string
	//go:embed query/billing_address_query_by_account_id.sql
	billingAddressQueryByAccountIDSQL string
	//go:embed query/billing_address_query_by_id.sql
	billingAddressQueryByIDSQL string
)

// Store manages the set of APIs for billing address database access.
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
func (s *Store) NewWithTx(tx pgsql.CommitRollbacker) (billing_address2.Storer, error) {
	ec, err := pgsql.GetExtContext(tx)
	if err != nil {
		return nil, fmt.Errorf("billing address transaction init error: %w", err)
	}

	store := Store{
		log:    s.log,
		dbPool: ec,
	}

	return &store, nil
}

// Create inserts a new billing address into the database.
func (s *Store) Create(ctx context.Context, addr billing_address2.BillingAddress) error {
	err := pgsql.NamedExecContext(
		ctx,
		s.log,
		s.dbPool,
		billingAddressCreateSQL,
		toBillingAddressDB(addr),
	)
	if err != nil {
		return fmt.Errorf("named_exec_context: %w", err)
	}

	return nil
}

// Update replaces a billing address in the database.
func (s *Store) Update(ctx context.Context, addr billing_address2.BillingAddress) error {
	err := pgsql.NamedExecContext(
		ctx,
		s.log,
		s.dbPool,
		billingAddressUpdateSQL,
		toBillingAddressDB(addr),
	)
	if err != nil {
		return fmt.Errorf("named_exec_context: %w", err)
	}

	return nil
}

// Delete removes a billing address from the database.
func (s *Store) Delete(ctx context.Context, addr billing_address2.BillingAddress) error {
	err := pgsql.NamedExecContext(
		ctx,
		s.log,
		s.dbPool,
		billingAddressDeleteSQL,
		toBillingAddressDB(addr),
	)
	if err != nil {
		return fmt.Errorf("named_exec_context: %w", err)
	}

	return nil
}

// QueryByAccountID retrieves billing addresses for an account.
func (s *Store) QueryByAccountID(
	ctx context.Context,
	accountID uuid.UUID,
) ([]billing_address2.BillingAddress, error) {
	data := struct {
		AccountID string `db:"account_id"`
	}{
		AccountID: accountID.String(),
	}

	var dbAddrs []billingAddressDB

	err := pgsql.NamedQuerySlice(
		ctx,
		s.log,
		s.dbPool,
		billingAddressQueryByAccountIDSQL,
		data,
		&dbAddrs,
	)
	if err != nil {
		return nil, fmt.Errorf("named_query_slice: %w", err)
	}

	return toBillingAddressesDomain(dbAddrs), nil
}

// QueryByID retrieves a billing address by its ID.
func (s *Store) QueryByID(
	ctx context.Context,
	id uuid.UUID,
) (billing_address2.BillingAddress, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: id.String(),
	}

	var dbAddr billingAddressDB

	err := pgsql.NamedQueryStruct(ctx, s.log, s.dbPool, billingAddressQueryByIDSQL, data, &dbAddr)
	if err != nil {
		if errors.Is(err, pgsql.ErrDBNotFound) {
			return billing_address2.BillingAddress{}, fmt.Errorf(
				"db: %w",
				billing_address2.ErrNotFound,
			)
		}

		return billing_address2.BillingAddress{}, fmt.Errorf("db: %w", err)
	}

	return toBillingAddressDomain(dbAddr), nil
}
