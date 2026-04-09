package db

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/order"
	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// PermissionQueryFilter is the set of optional predicates QueryPermissions
// accepts. It lives in the db package (not the domain), so the Storer
// interface stays self-contained, and callers don't drag a domain type into
// the data layer.
type PermissionQueryFilter struct {
	ID   *uuid.UUID
	Name *string
}

// permissionOrderByFields whitelists order fields → column names. Squirrel
// does not escape ORDER BY, so the whitelist is mandatory.
var permissionOrderByFields = map[string]string{
	"id":   "id",
	"name": "name",
}

// QueryPermissions returns permissions matching filter, ordered and
// cursor-paginated. It is the dynamic counterpart to the sqlc-generated
// static permission queries and runs on the same DBTX (pool or tx) as the
// embedded *Queries.
func (s *Store) QueryPermissions(
	ctx context.Context,
	filter PermissionQueryFilter,
	orderBy order.By,
	cur cursor.Cursor,
) ([]Permission, error) {
	col, ok := permissionOrderByFields[orderBy.Field]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrOrderFieldNotFound, orderBy.Field)
	}

	sb := Builder.
		Select("id", "name", "created_at", "updated_at", "deleted_at").
		From("permissions").
		Where(permissionFilterPredicates(filter)).
		OrderBy(col+" "+orderBy.Direction, "id "+orderBy.Direction).
		Limit(uint64(cur.Limit() + 1))

	if cp := permissionCursorPredicate(cur, orderBy, col); cp != nil {
		sb = sb.Where(cp)
	}

	query, args, err := sb.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query permissions: %w", err)
	}
	defer rows.Close()

	perms, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (Permission, error) {
		var p Permission
		err := row.Scan(
			&p.ID,
			&p.Name,
			&p.CreatedAt,
			&p.UpdatedAt,
			&p.DeletedAt,
		)
		return p, err
	})
	if err != nil {
		return nil, fmt.Errorf("collect rows: %w", err)
	}

	return perms, nil
}

func permissionFilterPredicates(f PermissionQueryFilter) sq.Sqlizer {
	preds := sq.And{sq.Eq{"deleted_at": nil}}

	if f.ID != nil {
		preds = append(preds, sq.Eq{"id": *f.ID})
	}

	if f.Name != nil {
		preds = append(preds, sq.Like{"name": "%" + *f.Name + "%"})
	}

	return preds
}

func permissionCursorPredicate(cur cursor.Cursor, orderBy order.By, col string) sq.Sqlizer {
	if !cur.HasCursor() {
		return nil
	}

	op := ">"
	if orderBy.Direction == order.DESC {
		op = "<"
	}

	return sq.Expr(fmt.Sprintf("(%s, id) %s (?, ?)", col, op), cur.FieldValue(), cur.ID())
}
