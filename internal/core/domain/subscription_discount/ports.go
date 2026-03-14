package subscription_discount

import (
	"context"

	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
)

type Storer interface {
	NewWithTx(tx pgsql.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, sd SubscriptionDiscount) error
	Query(
		ctx context.Context,
		filter QueryFilter,
		orderBy order.By,
		cur cursor.Cursor,
	) ([]SubscriptionDiscount, error)
	QueryByID(ctx context.Context, subscriptionDiscountID uuid.UUID) (SubscriptionDiscount, error)
	QueryBySubscriptionID(
		ctx context.Context,
		subscriptionID uuid.UUID,
	) ([]SubscriptionDiscount, error)
}
