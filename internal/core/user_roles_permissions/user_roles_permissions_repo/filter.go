package user_roles_permissions_repo

import (
	"fmt"

	urp "github.com/Housiadas/cerberus/internal/core/user_roles_permissions"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/order"
	sq "github.com/Masterminds/squirrel"
)

// orderByFields whitelists the domain order fields and maps them to their
// underlying column. Squirrel does not escape ORDER BY, so the whitelist is
// mandatory.
var orderByFields = map[string]string{
	urp.OrderByUserName:       "user_name",
	urp.OrderByUserEmail:      "user_email",
	urp.OrderByRoleName:       "role_name",
	urp.OrderByPermissionName: "permission_name",
}

func filterPredicates(f urp.QueryFilter) sq.Sqlizer {
	preds := sq.And{}

	if f.UserID != nil {
		preds = append(preds, sq.Eq{"user_id": *f.UserID})
	}

	if f.UserName != nil {
		preds = append(preds, sq.Like{"user_name": "%" + f.UserName.String() + "%"})
	}

	if f.UserEmail != nil {
		preds = append(preds, sq.Eq{"user_email": f.UserEmail.Address})
	}

	if f.RoleID != nil {
		preds = append(preds, sq.Eq{"role_id": *f.RoleID})
	}

	if f.RoleName != nil {
		preds = append(preds, sq.Like{"role_name": "%" + f.RoleName.String() + "%"})
	}

	if f.PermissionID != nil {
		preds = append(preds, sq.Eq{"permission_id": *f.PermissionID})
	}

	if f.PermissionName != nil {
		preds = append(preds, sq.Like{"permission_name": "%" + f.PermissionName.String() + "%"})
	}

	return preds
}

func cursorPredicate(cur cursor.Cursor, ob order.By) sq.Sqlizer {
	if !cur.HasCursor() {
		return nil
	}

	col, ok := orderByFields[ob.Field]
	if !ok {
		return nil
	}

	op := ">"
	if ob.Direction == order.DESC {
		op = "<"
	}

	return sq.Expr(fmt.Sprintf("(%s, user_id) %s (?, ?)", col, op), cur.FieldValue(), cur.ID())
}
