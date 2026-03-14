// Package subscription_discount_repo contains database-related CRUD functionality.
package subscription_discount_repo

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/subscription_discount"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// queries.
var (
	//go:embed query/subscription_discount_create.sql
	subscriptionDiscountCreateSQL string
	//go:embed query/subscription_discount_query.sql
	subscriptionDiscountQuerySQL string
	//go:embed query/subscription_discount_query_by_id.sql
	subscriptionDiscountQueryByIDSQL string
	//go:embed query/subscription_discount_query_by_subscription_id.sql
	subscriptionDiscountQueryBySubscriptionIDSQL string
)

// Store manages the set of APIs for subscription_discount database access.
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
func (s *Store) NewWithTx(tx pgsql.CommitRollbacker) (subscription_discount.Storer, error) {
	ec, err := pgsql.GetExtContext(tx)
	if err != nil {
		return nil, fmt.Errorf("subscription discount transaction init error: %w", err)
	}

	store := Store{
		log: s.log,
		db:  ec,
	}

	return &store, nil
}

// Create inserts a new subscription discount into the database.
func (s *Store) Create(ctx context.Context, sd subscription_discount.SubscriptionDiscount) error {
	err := pgsql.NamedExecContext(ctx, s.log, s.db, subscriptionDiscountCreateSQL, toSubscriptionDiscountDB(sd))
	if err != nil {
		return fmt.Errorf("error subscription discount create in db: %w", err)
	}

	return nil
}

// QueryByID gets the specified subscription discount from the database.
func (s *Store) QueryByID(ctx context.Context, subscriptionDiscountID uuid.UUID) (subscription_discount.SubscriptionDiscount, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: subscriptionDiscountID.String(),
	}

	var dbSD subscriptionDiscountDB

	err := pgsql.NamedQueryStruct(ctx, s.log, s.db, subscriptionDiscountQueryByIDSQL, data, &dbSD)
	if err != nil {
		if errors.Is(err, pgsql.ErrDBNotFound) {
			return subscription_discount.SubscriptionDiscount{}, fmt.Errorf("db: %w", subscription_discount.ErrNotFound)
		}

		return subscription_discount.SubscriptionDiscount{}, fmt.Errorf("db: %w", err)
	}

	return toSubscriptionDiscountDomain(dbSD), nil
}

// QueryBySubscriptionID gets subscription discounts by subscription ID from the database.
func (s *Store) QueryBySubscriptionID(ctx context.Context, subscriptionID uuid.UUID) ([]subscription_discount.SubscriptionDiscount, error) {
	data := struct {
		SubscriptionID string `db:"subscription_id"`
	}{
		SubscriptionID: subscriptionID.String(),
	}

	var dbSDs []subscriptionDiscountDB

	err := pgsql.NamedQuerySlice(ctx, s.log, s.db, subscriptionDiscountQueryBySubscriptionIDSQL, data, &dbSDs)
	if err != nil {
		return nil, fmt.Errorf("error query subscription discount by subscription id in db: %w", err)
	}

	return toSubscriptionDiscountsDomain(dbSDs), nil
}

// Query retrieves a list of existing subscription discounts from the database.
func (s *Store) Query(
	ctx context.Context,
	filter subscription_discount.QueryFilter,
	orderBy order.By,
	cur cursor.Cursor,
) ([]subscription_discount.SubscriptionDiscount, error) {
	data := map[string]any{
		"limit": cur.Limit() + 1,
	}

	buf := bytes.NewBufferString(subscriptionDiscountQuerySQL)
	applyFilter(filter, data, buf)
	applyCursor(cur, orderBy, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" FETCH NEXT :limit ROWS ONLY")

	var dbSDs []subscriptionDiscountDB

	err = pgsql.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbSDs)
	if err != nil {
		return nil, fmt.Errorf("error query subscription discount in db: %w", err)
	}

	return toSubscriptionDiscountsDomain(dbSDs), nil
}
