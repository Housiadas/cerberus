// Package subscription_repo contains database-related CRUD functionality.
package subscription_repo

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/subscription"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// queries.
var (
	//go:embed query/subscription_create.sql
	subscriptionCreateSQL string
	//go:embed query/subscription_update.sql
	subscriptionUpdateSQL string
	//go:embed query/subscription_query.sql
	subscriptionQuerySQL string
	//go:embed query/subscription_query_by_id.sql
	subscriptionQueryByIDSQL string
	//go:embed query/subscription_query_by_account_id.sql
	subscriptionQueryByAccountIDSQL string
)

// Store manages the set of APIs for subscriptionDB database access.
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
func (s *Store) NewWithTx(tx pgsql.CommitRollbacker) (subscription.Storer, error) {
	ec, err := pgsql.GetExtContext(tx)
	if err != nil {
		return nil, fmt.Errorf("subscription transaction init error: %w", err)
	}

	store := Store{
		log: s.log,
		db:  ec,
	}

	return &store, nil
}

// Create inserts a new subscription into the database.
func (s *Store) Create(ctx context.Context, sub subscription.Subscription) error {
	err := pgsql.NamedExecContext(ctx, s.log, s.db, subscriptionCreateSQL, toSubscriptionDB(sub))
	if err != nil {
		return fmt.Errorf("error subscription create in db: %w", err)
	}

	return nil
}

// Update replaces a subscription document in the database.
func (s *Store) Update(ctx context.Context, sub subscription.Subscription) error {
	err := pgsql.NamedExecContext(ctx, s.log, s.db, subscriptionUpdateSQL, toSubscriptionDB(sub))
	if err != nil {
		return fmt.Errorf("error subscription update in db: %w", err)
	}

	return nil
}

// QueryByID gets the specified subscription from the database.
func (s *Store) QueryByID(ctx context.Context, subscriptionID uuid.UUID) (subscription.Subscription, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: subscriptionID.String(),
	}

	var dbSub subscriptionDB

	err := pgsql.NamedQueryStruct(ctx, s.log, s.db, subscriptionQueryByIDSQL, data, &dbSub)
	if err != nil {
		if errors.Is(err, pgsql.ErrDBNotFound) {
			return subscription.Subscription{}, fmt.Errorf("db: %w", subscription.ErrNotFound)
		}

		return subscription.Subscription{}, fmt.Errorf("db: %w", err)
	}

	return toDomain(dbSub)
}

// QueryByAccountID retrieves subscriptions for a specific account from the database.
func (s *Store) QueryByAccountID(ctx context.Context, accountID uuid.UUID) ([]subscription.Subscription, error) {
	data := struct {
		AccountID string `db:"account_id"`
	}{
		AccountID: accountID.String(),
	}

	var dbSubs []subscriptionDB

	err := pgsql.NamedQuerySlice(ctx, s.log, s.db, subscriptionQueryByAccountIDSQL, data, &dbSubs)
	if err != nil {
		return nil, fmt.Errorf("error query subscription by account id in db: %w", err)
	}

	return toSliceDomain(dbSubs)
}

// Query retrieves a list of existing subscriptions from the database.
func (s *Store) Query(
	ctx context.Context,
	filter subscription.QueryFilter,
	orderBy order.By,
	cur cursor.Cursor,
) ([]subscription.Subscription, error) {
	data := map[string]any{
		"limit": cur.Limit() + 1,
	}

	buf := bytes.NewBufferString(subscriptionQuerySQL)
	applyFilter(filter, data, buf)
	applyCursor(cur, orderBy, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" FETCH NEXT :limit ROWS ONLY")

	var dbSubs []subscriptionDB

	err = pgsql.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbSubs)
	if err != nil {
		return nil, fmt.Errorf("error query subscription in db: %w", err)
	}

	return toSliceDomain(dbSubs)
}
