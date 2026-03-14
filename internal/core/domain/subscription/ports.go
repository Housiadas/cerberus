package subscription

import (
	"context"

	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
)

type Storer interface {
	NewWithTx(tx pgsql.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, sub Subscription) error
	Update(ctx context.Context, sub Subscription) error
	Query(
		ctx context.Context,
		filter QueryFilter,
		orderBy order.By,
		cur cursor.Cursor,
	) ([]Subscription, error)
	QueryByID(ctx context.Context, subscriptionID uuid.UUID) (Subscription, error)
	QueryByAccountID(ctx context.Context, accountID uuid.UUID) ([]Subscription, error)
}
