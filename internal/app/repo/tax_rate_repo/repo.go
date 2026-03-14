// Package tax_rate_repo contains database-related CRUD functionality.
package tax_rate_repo

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/tax_rate"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// queries.
var (
	//go:embed query/tax_rate_create.sql
	taxRateCreateSQL string
	//go:embed query/tax_rate_update.sql
	taxRateUpdateSQL string
	//go:embed query/tax_rate_delete.sql
	taxRateDeleteSQL string
	//go:embed query/tax_rate_query.sql
	taxRateQuerySQL string
	//go:embed query/tax_rate_query_by_id.sql
	taxRateQueryByIDSQL string
)

// Store manages the set of APIs for tax rate database access.
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
func (s *Store) NewWithTx(tx pgsql.CommitRollbacker) (tax_rate.Storer, error) {
	ec, err := pgsql.GetExtContext(tx)
	if err != nil {
		return nil, fmt.Errorf("tax rate transaction init error: %w", err)
	}

	store := Store{
		log: s.log,
		db:  ec,
	}

	return &store, nil
}

// Create inserts a new tax rate into the database.
func (s *Store) Create(ctx context.Context, tr tax_rate.TaxRate) error {
	err := pgsql.NamedExecContext(ctx, s.log, s.db, taxRateCreateSQL, toTaxRateDB(tr))
	if err != nil {
		return fmt.Errorf("error tax rate create in db: %w", err)
	}

	return nil
}

// Update replaces a tax rate document in the database.
func (s *Store) Update(ctx context.Context, tr tax_rate.TaxRate) error {
	err := pgsql.NamedExecContext(ctx, s.log, s.db, taxRateUpdateSQL, toTaxRateDB(tr))
	if err != nil {
		return fmt.Errorf("error tax rate update in db: %w", err)
	}

	return nil
}

// Delete removes a tax rate from the database.
func (s *Store) Delete(ctx context.Context, tr tax_rate.TaxRate) error {
	err := pgsql.NamedExecContext(ctx, s.log, s.db, taxRateDeleteSQL, toTaxRateDB(tr))
	if err != nil {
		return fmt.Errorf("error delete tax rate in db: %w", err)
	}

	return nil
}

// QueryByID gets the specified tax rate from the database.
func (s *Store) QueryByID(ctx context.Context, taxRateID uuid.UUID) (tax_rate.TaxRate, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: taxRateID.String(),
	}

	var dbTaxRate taxRateDB

	err := pgsql.NamedQueryStruct(ctx, s.log, s.db, taxRateQueryByIDSQL, data, &dbTaxRate)
	if err != nil {
		if errors.Is(err, pgsql.ErrDBNotFound) {
			return tax_rate.TaxRate{}, fmt.Errorf("db: %w", tax_rate.ErrNotFound)
		}

		return tax_rate.TaxRate{}, fmt.Errorf("db: %w", err)
	}

	return toTaxRateDomain(dbTaxRate)
}

// Query retrieves a list of existing tax rates from the database.
func (s *Store) Query(
	ctx context.Context,
	filter tax_rate.QueryFilter,
	orderBy order.By,
	cur cursor.Cursor,
) ([]tax_rate.TaxRate, error) {
	data := map[string]any{
		"limit": cur.Limit() + 1,
	}

	buf := bytes.NewBufferString(taxRateQuerySQL)
	applyFilter(filter, data, buf)
	applyCursor(cur, orderBy, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" FETCH NEXT :limit ROWS ONLY")

	var dbTaxRates []taxRateDB

	err = pgsql.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbTaxRates)
	if err != nil {
		return nil, fmt.Errorf("error query tax rate in db: %w", err)
	}

	return toTaxRatesDomain(dbTaxRates)
}
