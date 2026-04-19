package db //nolint:dupl // role and permission queries are structurally similar but type-distinct

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/order"
	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// RoleQueryFilter is the set of optional predicates QueryRoles accepts.
// It lives in the db package (not the domain), so the Storer interface stays
// self-contained, and callers don't drag a domain type into the data layer.
type RoleQueryFilter struct {
	ID   *uuid.UUID
	Name *string
}

// roleOrderByFields whitelists order fields → column names. Squirrel does
// not escape ORDER BY, so the whitelist is mandatory.
var roleOrderByFields = map[string]string{
	"id":   "id",
	"name": "name",
}

// QueryRoles returns roles matching filter, ordered and cursor-paginated.
// It is the dynamic counterpart to the sqlc-generated static role queries
// and runs on the same DBTX (pool or tx) as the embedded *Queries.
func (s *Store) QueryRoles(
	ctx context.Context,
	filter RoleQueryFilter,
	orderBy order.By,
	cur cursor.Cursor,
) ([]Role, error) {
	col, ok := roleOrderByFields[orderBy.Field]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrOrderFieldNotFound, orderBy.Field)
	}

	sb := Builder.
		Select("id", "name", "created_at", "updated_at", "deleted_at").
		From("roles").
		Where(roleFilterPredicates(filter)).
		OrderBy(col+" "+orderBy.Direction, "id "+orderBy.Direction).
		Limit(uint64(cur.Limit() + 1))

	if cp := cursorPredicate(cur, orderBy, col); cp != nil {
		sb = sb.Where(cp)
	}

	query, args, err := sb.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query roles: %w", err)
	}
	defer rows.Close()

	roles, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (Role, error) {
		var r Role

		err := row.Scan(
			&r.ID,
			&r.Name,
			&r.CreatedAt,
			&r.UpdatedAt,
			&r.DeletedAt,
		)
		if err != nil {
			return r, fmt.Errorf("scan row: %w", err)
		}

		return r, nil
	})
	if err != nil {
		return nil, fmt.Errorf("collect rows: %w", err)
	}

	return roles, nil
}

func roleFilterPredicates(f RoleQueryFilter) sq.Sqlizer {
	preds := sq.And{sq.Eq{"deleted_at": nil}}

	if f.ID != nil {
		preds = append(preds, sq.Eq{"id": *f.ID})
	}

	if f.Name != nil {
		preds = append(preds, sq.Like{"name": "%" + *f.Name + "%"})
	}

	return preds
}
