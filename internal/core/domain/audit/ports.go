package audit

import (
	"context"

	"github.com/Housiadas/cerberus/internal/utils/page"
	"github.com/Housiadas/cerberus/pkg/order"
)

// Storer interface declares the behavior this package needs to persist and retrieve data.
type Storer interface {
	Create(ctx context.Context, audit Audit) error
	Query(
		ctx context.Context,
		filter QueryFilter,
		orderBy order.By,
		page page.Page,
	) ([]Audit, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
}
