package audit

import (
	"context"
	"time"

	db "github.com/Housiadas/cerberus/db/sqlc"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/order"
)

type clock interface {
	Now() time.Time
}

// store interface declares the behavior this package needs to persist and retrieve data.
type storer interface {
	CreateAudit(ctx context.Context, arg db.CreateAuditParams) (db.Audit, error)
	QueryAudits(
		ctx context.Context,
		filter db.AuditQueryFilter,
		orderBy order.By,
		cur cursor.Cursor,
	) ([]db.Audit, error)
}
