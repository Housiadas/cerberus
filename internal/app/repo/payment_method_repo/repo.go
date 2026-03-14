// Package payment_method_repo contains database-related CRUD functionality.
package payment_method_repo

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/payment_method"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// queries.
var (
	//go:embed query/payment_method_create.sql
	paymentMethodCreateSQL string
	//go:embed query/payment_method_delete.sql
	paymentMethodDeleteSQL string
	//go:embed query/payment_method_query.sql
	paymentMethodQuerySQL string
	//go:embed query/payment_method_query_by_id.sql
	paymentMethodQueryByIDSQL string
	//go:embed query/payment_method_query_by_account_id.sql
	paymentMethodQueryByAccountIDSQL string
)

// Store manages the set of APIs for paymentMethodDB database access.
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
func (s *Store) NewWithTx(tx pgsql.CommitRollbacker) (payment_method.Storer, error) {
	ec, err := pgsql.GetExtContext(tx)
	if err != nil {
		return nil, fmt.Errorf("payment method transaction init error: %w", err)
	}

	store := Store{
		log: s.log,
		db:  ec,
	}

	return &store, nil
}

// Create inserts a new payment method into the database.
func (s *Store) Create(ctx context.Context, pm payment_method.PaymentMethod) error {
	err := pgsql.NamedExecContext(ctx, s.log, s.db, paymentMethodCreateSQL, toPaymentMethodDB(pm))
	if err != nil {
		return fmt.Errorf("error payment method create in db: %w", err)
	}

	return nil
}

// Delete removes a payment method from the database.
func (s *Store) Delete(ctx context.Context, pm payment_method.PaymentMethod) error {
	data := struct {
		ID string `db:"id"`
	}{
		ID: pm.ID().String(),
	}

	err := pgsql.NamedExecContext(ctx, s.log, s.db, paymentMethodDeleteSQL, data)
	if err != nil {
		return fmt.Errorf("error delete payment method in db: %w", err)
	}

	return nil
}

// QueryByID gets the specified payment method from the database.
func (s *Store) QueryByID(
	ctx context.Context,
	paymentMethodID uuid.UUID,
) (payment_method.PaymentMethod, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: paymentMethodID.String(),
	}

	var dbPM paymentMethodDB

	err := pgsql.NamedQueryStruct(ctx, s.log, s.db, paymentMethodQueryByIDSQL, data, &dbPM)
	if err != nil {
		if errors.Is(err, pgsql.ErrDBNotFound) {
			return payment_method.PaymentMethod{}, fmt.Errorf("db: %w", payment_method.ErrNotFound)
		}

		return payment_method.PaymentMethod{}, fmt.Errorf("db: %w", err)
	}

	return toDomain(dbPM)
}

// QueryByAccountID retrieves payment methods for a specific account from the database.
func (s *Store) QueryByAccountID(
	ctx context.Context,
	accountID uuid.UUID,
) ([]payment_method.PaymentMethod, error) {
	data := struct {
		AccountID string `db:"account_id"`
	}{
		AccountID: accountID.String(),
	}

	var dbPMs []paymentMethodDB

	err := pgsql.NamedQuerySlice(ctx, s.log, s.db, paymentMethodQueryByAccountIDSQL, data, &dbPMs)
	if err != nil {
		return nil, fmt.Errorf("error query payment method by account id in db: %w", err)
	}

	return toSliceDomain(dbPMs)
}

// Query retrieves a list of existing payment methods from the database.
func (s *Store) Query(
	ctx context.Context,
	filter payment_method.QueryFilter,
	orderBy order.By,
	cur cursor.Cursor,
) ([]payment_method.PaymentMethod, error) {
	data := map[string]any{
		"limit": cur.Limit() + 1,
	}

	buf := bytes.NewBufferString(paymentMethodQuerySQL)
	applyFilter(filter, data, buf)
	applyCursor(cur, orderBy, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" FETCH NEXT :limit ROWS ONLY")

	var dbPMs []paymentMethodDB

	err = pgsql.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbPMs)
	if err != nil {
		return nil, fmt.Errorf("error query payment method in db: %w", err)
	}

	return toSliceDomain(dbPMs)
}
