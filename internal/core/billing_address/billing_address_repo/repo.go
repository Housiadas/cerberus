// Package billing_address_repo contains database related CRUD functionality for billing addresses.
package billing_address_repo

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/billing_address"
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
	log *logger.Service
	db  *sqlx.DB
}

// NewStore constructs the api for data access.
func NewStore(log *logger.Service, db *sqlx.DB) *Store {
	return &Store{
		log: log,
		db:  db,
	}
}

// Create inserts a new billing address into the database.
func (s *Store) Create(ctx context.Context, addr billing_address.BillingAddress) error {
	err := pgsql.NamedExecContext(
		ctx,
		s.log,
		pgsql.Conn(ctx, s.db),
		billingAddressCreateSQL,
		toBillingAddressDB(addr),
	)
	if err != nil {
		return fmt.Errorf("named_exec_context: %w", err)
	}

	return nil
}

// Update replaces a billing address in the database.
func (s *Store) Update(ctx context.Context, addr billing_address.BillingAddress) error {
	err := pgsql.NamedExecContext(
		ctx,
		s.log,
		pgsql.Conn(ctx, s.db),
		billingAddressUpdateSQL,
		toBillingAddressDB(addr),
	)
	if err != nil {
		return fmt.Errorf("named_exec_context: %w", err)
	}

	return nil
}

// Delete removes a billing address from the database.
func (s *Store) Delete(ctx context.Context, addr billing_address.BillingAddress) error {
	err := pgsql.NamedExecContext(
		ctx,
		s.log,
		pgsql.Conn(ctx, s.db),
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
) ([]billing_address.BillingAddress, error) {
	data := struct {
		AccountID string `db:"account_id"`
	}{
		AccountID: accountID.String(),
	}

	var dbAddrs []billingAddressDB

	err := pgsql.NamedQuerySlice(
		ctx,
		s.log,
		pgsql.Conn(ctx, s.db),
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
) (billing_address.BillingAddress, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: id.String(),
	}

	var dbAddr billingAddressDB

	err := pgsql.NamedQueryStruct(ctx, s.log, pgsql.Conn(ctx, s.db), billingAddressQueryByIDSQL, data, &dbAddr)
	if err != nil {
		if errors.Is(err, pgsql.ErrDBNotFound) {
			return billing_address.BillingAddress{}, fmt.Errorf(
				"db: %w",
				billing_address.ErrNotFound,
			)
		}

		return billing_address.BillingAddress{}, fmt.Errorf("db: %w", err)
	}

	return toBillingAddressDomain(dbAddr), nil
}
