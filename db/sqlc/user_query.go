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

// UserQueryFilter is the set of optional predicates QueryUsers accepts.
// It lives in the db package (not the domain), so the Storer interface stays
// self-contained, and callers don't drag a domain type into the data layer.
type UserQueryFilter struct {
	ID               *uuid.UUID
	Name             *string
	Email            *string
	StartCreatedDate *time.Time
	EndCreatedDate   *time.Time
}

// userOrderByFields whitelists order fields → column names. Squirrel does
// not escape ORDER BY, so the whitelist is mandatory.
var userOrderByFields = map[string]string{
	"id":      "id",
	"name":    "name",
	"email":   "email",
	"enabled": "enabled",
}

// QueryUsers returns users matching filter, ordered and cursor-paginated.
// It is the dynamic counterpart to the sqlc-generated static user queries
// and runs on the same DBTX (pool or tx) as the embedded *Queries.
func (s *Store) QueryUsers(
	ctx context.Context,
	filter UserQueryFilter,
	orderBy order.By,
	cur cursor.Cursor,
) ([]User, error) {
	col, ok := userOrderByFields[orderBy.Field]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrOrderFieldNotFound, orderBy.Field)
	}

	sb := Builder.
		Select(
			"id", "account_id", "name", "email", "password_hash",
			"department", "enabled", "created_at", "updated_at", "deleted_at",
		).
		From("users").
		Where(userFilterPredicates(filter)).
		OrderBy(col+" "+orderBy.Direction, "id "+orderBy.Direction).
		Limit(uint64(cur.Limit() + 1))

	if cp := userCursorPredicate(cur, orderBy, col); cp != nil {
		sb = sb.Where(cp)
	}

	query, args, err := sb.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query users: %w", err)
	}
	defer rows.Close()

	users, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (User, error) {
		var u User
		err := row.Scan(
			&u.ID,
			&u.AccountID,
			&u.Name,
			&u.Email,
			&u.PasswordHash,
			&u.Department,
			&u.Enabled,
			&u.CreatedAt,
			&u.UpdatedAt,
			&u.DeletedAt,
		)
		return u, err
	})
	if err != nil {
		return nil, fmt.Errorf("collect rows: %w", err)
	}

	return users, nil
}

func userFilterPredicates(f UserQueryFilter) sq.Sqlizer {
	preds := sq.And{sq.Eq{"deleted_at": nil}}

	if f.ID != nil {
		preds = append(preds, sq.Eq{"id": *f.ID})
	}

	if f.Name != nil {
		preds = append(preds, sq.Like{"name": "%" + *f.Name + "%"})
	}

	if f.Email != nil {
		preds = append(preds, sq.Eq{"email": *f.Email})
	}

	if f.StartCreatedDate != nil {
		preds = append(preds, sq.GtOrEq{"created_at": f.StartCreatedDate.UTC()})
	}

	if f.EndCreatedDate != nil {
		preds = append(preds, sq.LtOrEq{"created_at": f.EndCreatedDate.UTC()})
	}

	return preds
}

func userCursorPredicate(cur cursor.Cursor, orderBy order.By, col string) sq.Sqlizer {
	if !cur.HasCursor() {
		return nil
	}

	op := ">"
	if orderBy.Direction == order.DESC {
		op = "<"
	}

	return sq.Expr(fmt.Sprintf("(%s, id) %s (?, ?)", col, op), cur.FieldValue(), cur.ID())
}
