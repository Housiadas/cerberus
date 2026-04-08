package user_repo

import (
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/user"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/order"
	sq "github.com/Masterminds/squirrel"
)

// orderByFields whitelists the domain order fields and maps them to their
// underlying column. Squirrel does not escape ORDER BY, so the whitelist is
// mandatory.
var orderByFields = map[string]string{
	user.OrderByID:      "id",
	user.OrderByName:    "name",
	user.OrderByEmail:   "email",
	user.OrderByEnabled: "enabled",
}

func filterPredicates(f user.QueryFilter) sq.Sqlizer {
	preds := sq.And{sq.Eq{"deleted_at": nil}}

	if f.ID != nil {
		preds = append(preds, sq.Eq{"id": *f.ID})
	}

	if f.Name != nil {
		preds = append(preds, sq.Like{"name": "%" + f.Name.String() + "%"})
	}

	if f.Email != nil {
		preds = append(preds, sq.Eq{"email": (*f.Email).String()})
	}

	if f.StartCreatedDate != nil {
		preds = append(preds, sq.GtOrEq{"created_at": f.StartCreatedDate.UTC()})
	}

	if f.EndCreatedDate != nil {
		preds = append(preds, sq.LtOrEq{"created_at": f.EndCreatedDate.UTC()})
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
