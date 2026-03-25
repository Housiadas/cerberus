// Package subscription_repo contains database related CRUD functionality for subscriptions.
package subscription_repo

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	subscription2 "github.com/Housiadas/cerberus/internal/core/subscription"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var (
	//go:embed query/subscription_create.sql
	subscriptionCreateSQL string
	//go:embed query/subscription_update.sql
	subscriptionUpdateSQL string
	//go:embed query/subscription_query_by_account_id.sql
	subscriptionQueryByAccountIDSQL string
	//go:embed query/subscription_query_by_stripe_id.sql
	subscriptionQueryByStripeIDSQL string
)

// Store manages the set of APIs for subscription database access.
type Store struct {
	log    *logger.Service
	dbPool sqlx.ExtContext
}

// NewStore constructs the api for data access.
func NewStore(log *logger.Service, dbPool *sqlx.DB) *Store {
	return &Store{
		log:    log,
		dbPool: dbPool,
	}
}

// NewWithTx constructs a new Store value replacing the sqlx DB
// value with a sqlx DB value that is currently inside a transaction.
func (s *Store) NewWithTx(tx pgsql.CommitRollbacker) (subscription2.Storer, error) {
	ec, err := pgsql.GetExtContext(tx)
	if err != nil {
		return nil, fmt.Errorf("subscription transaction init error: %w", err)
	}

	store := Store{
		log:    s.log,
		dbPool: ec,
	}

	return &store, nil
}

// Create inserts a new subscription into the database.
func (s *Store) Create(ctx context.Context, sub subscription2.Subscription) error {
	err := pgsql.NamedExecContext(
		ctx,
		s.log,
		s.dbPool,
		subscriptionCreateSQL,
		toSubscriptionDB(sub),
	)
	if err != nil {
		return fmt.Errorf("named_exec_context: %w", err)
	}

	return nil
}

// Update replaces a subscription in the database.
func (s *Store) Update(ctx context.Context, sub subscription2.Subscription) error {
	err := pgsql.NamedExecContext(
		ctx,
		s.log,
		s.dbPool,
		subscriptionUpdateSQL,
		toSubscriptionDB(sub),
	)
	if err != nil {
		return fmt.Errorf("named_exec_context: %w", err)
	}

	return nil
}

// QueryByAccountID retrieves subscriptions for an account.
func (s *Store) QueryByAccountID(
	ctx context.Context,
	accountID uuid.UUID,
) ([]subscription2.Subscription, error) {
	data := struct {
		AccountID string `db:"account_id"`
	}{
		AccountID: accountID.String(),
	}

	var dbSubs []subscriptionDB

	err := pgsql.NamedQuerySlice(
		ctx,
		s.log,
		s.dbPool,
		subscriptionQueryByAccountIDSQL,
		data,
		&dbSubs,
	)
	if err != nil {
		return nil, fmt.Errorf("named_query_slice: %w", err)
	}

	return toSubscriptionsDomain(dbSubs), nil
}

// QueryByStripeID retrieves a subscription by its Stripe ID.
func (s *Store) QueryByStripeID(
	ctx context.Context,
	stripeSubID string,
) (subscription2.Subscription, error) {
	data := struct {
		StripeSubscriptionID string `db:"stripe_subscription_id"`
	}{
		StripeSubscriptionID: stripeSubID,
	}

	var dbSub subscriptionDB

	err := pgsql.NamedQueryStruct(
		ctx,
		s.log,
		s.dbPool,
		subscriptionQueryByStripeIDSQL,
		data,
		&dbSub,
	)
	if err != nil {
		if errors.Is(err, pgsql.ErrDBNotFound) {
			return subscription2.Subscription{}, fmt.Errorf("db: %w", subscription2.ErrNotFound)
		}

		return subscription2.Subscription{}, fmt.Errorf("db: %w", err)
	}

	return toSubscriptionDomain(dbSub), nil
}
