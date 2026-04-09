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

// AccountQueryFilter is the set of optional predicates QueryAccounts accepts.
// It lives in the db package (not the domain), so the Storer interface stays
// self-contained, and callers don't drag a domain type into the data layer.
type AccountQueryFilter struct {
	ID               *uuid.UUID
	Name             *string
	StripeCustomerID *string
}

// accountOrderByFields whitelists order fields → column names. Squirrel does
// not escape ORDER BY, so the whitelist is mandatory.
var accountOrderByFields = map[string]string{
	"id":   "id",
	"name": "name",
}

// QueryAccounts returns accounts matching filter, ordered and cursor-paginated.
// It is the dynamic counterpart to the sqlc-generated static account queries
// and runs on the same DBTX (pool or tx) as the embedded *Queries.
func (s *Store) QueryAccounts(
	ctx context.Context,
	filter AccountQueryFilter,
	orderBy order.By,
	cur cursor.Cursor,
) ([]Account, error) {
	col, ok := accountOrderByFields[orderBy.Field]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrOrderFieldNotFound, orderBy.Field)
	}

	sb := Builder.
		Select(
			"id", "name", "stripe_customer_id",
			"created_at", "updated_at", "deleted_at",
		).
		From("accounts").
		Where(accountFilterPredicates(filter)).
		OrderBy(col+" "+orderBy.Direction, "id "+orderBy.Direction).
		Limit(uint64(cur.Limit() + 1))

	if cp := accountCursorPredicate(cur, orderBy, col); cp != nil {
		sb = sb.Where(cp)
	}

	query, args, err := sb.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query accounts: %w", err)
	}
	defer rows.Close()

	accounts, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (Account, error) {
		var a Account
		err := row.Scan(
			&a.ID,
			&a.Name,
			&a.StripeCustomerID,
			&a.CreatedAt,
			&a.UpdatedAt,
			&a.DeletedAt,
		)
		return a, err
	})
	if err != nil {
		return nil, fmt.Errorf("collect rows: %w", err)
	}

	return accounts, nil
}

func accountFilterPredicates(f AccountQueryFilter) sq.Sqlizer {
	preds := sq.And{sq.Eq{"deleted_at": nil}}

	if f.ID != nil {
		preds = append(preds, sq.Eq{"id": *f.ID})
	}

	if f.Name != nil {
		preds = append(preds, sq.Like{"name": "%" + *f.Name + "%"})
	}

	if f.StripeCustomerID != nil {
		preds = append(preds, sq.Eq{"stripe_customer_id": *f.StripeCustomerID})
	}

	return preds
}

func accountCursorPredicate(cur cursor.Cursor, orderBy order.By, col string) sq.Sqlizer {
	if !cur.HasCursor() {
		return nil
	}

	op := ">"
	if orderBy.Direction == order.DESC {
		op = "<"
	}

	return sq.Expr(fmt.Sprintf("(%s, id) %s (?, ?)", col, op), cur.FieldValue(), cur.ID())
}
