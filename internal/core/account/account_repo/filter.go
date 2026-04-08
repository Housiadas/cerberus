package account_repo

import (
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/account"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/order"
	sq "github.com/Masterminds/squirrel"
)

// orderByFields whitelists the domain order fields and maps them to their
// underlying column. Squirrel does not escape ORDER BY, so the whitelist is
// mandatory.
var orderByFields = map[string]string{
	account.OrderByID:   "id",
	account.OrderByName: "name",
}

func filterPredicates(f account.QueryFilter) sq.Sqlizer {
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

func cursorPredicate(cur cursor.Cursor, orderBy order.By) sq.Sqlizer {
	if !cur.HasCursor() {
		return nil
	}

	col, ok := orderByFields[orderBy.Field]
	if !ok {
		return nil
	}

	op := ">"
	if orderBy.Direction == order.DESC {
		op = "<"
	}

	return sq.Expr(fmt.Sprintf("(%s, id) %s (?, ?)", col, op), cur.FieldValue(), cur.ID())
}
