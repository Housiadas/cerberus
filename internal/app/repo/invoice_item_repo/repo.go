// Package invoice_item_repo contains database-related CRUD functionality.
package invoice_item_repo

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/invoice_item"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// queries.
var (
	//go:embed query/invoice_item_create.sql
	invoiceItemCreateSQL string
	//go:embed query/invoice_item_query.sql
	invoiceItemQuerySQL string
	//go:embed query/invoice_item_query_by_id.sql
	invoiceItemQueryByIDSQL string
	//go:embed query/invoice_item_query_by_invoice_id.sql
	invoiceItemQueryByInvoiceIDSQL string
)

// Store manages the set of APIs for invoiceItemDB database access.
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
func (s *Store) NewWithTx(tx pgsql.CommitRollbacker) (invoice_item.Storer, error) {
	ec, err := pgsql.GetExtContext(tx)
	if err != nil {
		return nil, fmt.Errorf("invoice item transaction init error: %w", err)
	}

	store := Store{
		log: s.log,
		db:  ec,
	}

	return &store, nil
}

// Create inserts a new invoice item into the database.
func (s *Store) Create(ctx context.Context, item invoice_item.InvoiceItem) error {
	err := pgsql.NamedExecContext(ctx, s.log, s.db, invoiceItemCreateSQL, toInvoiceItemDB(item))
	if err != nil {
		return fmt.Errorf("error invoice item create in db: %w", err)
	}

	return nil
}

// QueryByID gets the specified invoice item from the database.
func (s *Store) QueryByID(ctx context.Context, invoiceItemID uuid.UUID) (invoice_item.InvoiceItem, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: invoiceItemID.String(),
	}

	var dbItem invoiceItemDB

	err := pgsql.NamedQueryStruct(ctx, s.log, s.db, invoiceItemQueryByIDSQL, data, &dbItem)
	if err != nil {
		if errors.Is(err, pgsql.ErrDBNotFound) {
			return invoice_item.InvoiceItem{}, fmt.Errorf("db: %w", invoice_item.ErrNotFound)
		}

		return invoice_item.InvoiceItem{}, fmt.Errorf("db: %w", err)
	}

	return toDomain(dbItem), nil
}

// QueryByInvoiceID retrieves invoice items for a specific invoice from the database.
func (s *Store) QueryByInvoiceID(ctx context.Context, invoiceID uuid.UUID) ([]invoice_item.InvoiceItem, error) {
	data := struct {
		InvoiceID string `db:"invoice_id"`
	}{
		InvoiceID: invoiceID.String(),
	}

	var dbItems []invoiceItemDB

	err := pgsql.NamedQuerySlice(ctx, s.log, s.db, invoiceItemQueryByInvoiceIDSQL, data, &dbItems)
	if err != nil {
		return nil, fmt.Errorf("error query invoice item by invoice id in db: %w", err)
	}

	return toSliceDomain(dbItems), nil
}

// Query retrieves a list of existing invoice items from the database.
func (s *Store) Query(
	ctx context.Context,
	filter invoice_item.QueryFilter,
	orderBy order.By,
	cur cursor.Cursor,
) ([]invoice_item.InvoiceItem, error) {
	data := map[string]any{
		"limit": cur.Limit() + 1,
	}

	buf := bytes.NewBufferString(invoiceItemQuerySQL)
	applyFilter(filter, data, buf)
	applyCursor(cur, orderBy, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" FETCH NEXT :limit ROWS ONLY")

	var dbItems []invoiceItemDB

	err = pgsql.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbItems)
	if err != nil {
		return nil, fmt.Errorf("error query invoice item in db: %w", err)
	}

	return toSliceDomain(dbItems), nil
}
