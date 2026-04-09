package db

import (
	"context"
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/order"
	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// AuditQueryFilter is the set of optional predicates QueryAudits accepts.
// It lives in the db package (not the domain), so the Storer interface stays
// self-contained, and callers don't drag a domain type into the data layer.
type AuditQueryFilter struct {
	ObjID     *uuid.UUID
	ObjEntity *string
	ObjName   *string
	ActorID   *uuid.UUID
	Action    *string
	Since     *time.Time
	Until     *time.Time
}

// auditOrderByFields whitelists order fields → column names. Squirrel does
// not escape ORDER BY, so the whitelist is mandatory. Keys are column names;
// callers must translate domain order keys before calling.
var auditOrderByFields = map[string]string{
	"obj_id":     "obj_id",
	"obj_entity": "obj_entity",
	"obj_name":   "obj_name",
	"actor_id":   "actor_id",
	"action":     "action",
}

// QueryAudits returns audit records matching filter, ordered and
// cursor-paginated. It is the dynamic counterpart to the sqlc-generated
// static audit queries and runs on the same DBTX (pool or tx) as the
// embedded *Queries.
func (s *Store) QueryAudits(
	ctx context.Context,
	filter AuditQueryFilter,
	orderBy order.By,
	cur cursor.Cursor,
) ([]Audit, error) {
	col, ok := auditOrderByFields[orderBy.Field]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrOrderFieldNotFound, orderBy.Field)
	}

	sb := Builder.
		Select(
			"id", "obj_id", "obj_entity", "obj_name",
			"actor_id", "action", "data", "message", "created_at",
		).
		From("audit").
		Where(auditFilterPredicates(filter)).
		OrderBy(col+" "+orderBy.Direction, "id "+orderBy.Direction).
		Limit(uint64(cur.Limit() + 1))

	if cp := auditCursorPredicate(cur, orderBy, col); cp != nil {
		sb = sb.Where(cp)
	}

	query, args, err := sb.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query audits: %w", err)
	}
	defer rows.Close()

	audits, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (Audit, error) {
		var a Audit
		err := row.Scan(
			&a.ID,
			&a.ObjID,
			&a.ObjEntity,
			&a.ObjName,
			&a.ActorID,
			&a.Action,
			&a.Data,
			&a.Message,
			&a.CreatedAt,
		)
		return a, err
	})
	if err != nil {
		return nil, fmt.Errorf("collect rows: %w", err)
	}

	return audits, nil
}

func auditFilterPredicates(f AuditQueryFilter) sq.Sqlizer {
	preds := sq.And{}

	if f.ObjID != nil {
		preds = append(preds, sq.Eq{"obj_id": *f.ObjID})
	}

	if f.ObjEntity != nil {
		preds = append(preds, sq.Eq{"obj_entity": *f.ObjEntity})
	}

	if f.ObjName != nil {
		preds = append(preds, sq.Like{"obj_name": "%" + *f.ObjName + "%"})
	}

	if f.ActorID != nil {
		preds = append(preds, sq.Eq{"actor_id": *f.ActorID})
	}

	if f.Action != nil {
		preds = append(preds, sq.Eq{"action": *f.Action})
	}

	if f.Since != nil {
		preds = append(preds, sq.GtOrEq{"created_at": f.Since.UTC()})
	}

	if f.Until != nil {
		preds = append(preds, sq.LtOrEq{"created_at": f.Until.UTC()})
	}

	return preds
}

func auditCursorPredicate(cur cursor.Cursor, orderBy order.By, col string) sq.Sqlizer {
	if !cur.HasCursor() {
		return nil
	}

	op := ">"
	if orderBy.Direction == order.DESC {
		op = "<"
	}

	return sq.Expr(fmt.Sprintf("(%s, id) %s (?, ?)", col, op), cur.FieldValue(), cur.ID())
}
