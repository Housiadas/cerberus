package audit_repo

import (
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/audit"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/order"
	sq "github.com/Masterminds/squirrel"
)

// orderByFields whitelists the domain order fields and maps them to their
// underlying column. Squirrel does not escape ORDER BY, so the whitelist is
// mandatory.
var orderByFields = map[string]string{
	audit.OrderByObjID:     "obj_id",
	audit.OrderByObjEntity: "obj_entity",
	audit.OrderByObjName:   "obj_name",
	audit.OrderByActorID:   "actor_id",
	audit.OrderByAction:    "action",
}

func filterPredicates(f audit.QueryFilter) sq.Sqlizer {
	preds := sq.And{}

	if f.ObjID != nil {
		preds = append(preds, sq.Eq{"obj_id": f.ObjID})
	}

	if f.ObjEntity != nil {
		preds = append(preds, sq.Eq{"obj_entity": f.ObjEntity.String()})
	}

	if f.ObjName != nil {
		preds = append(preds, sq.Like{"obj_name": "%" + f.ObjName.String() + "%"})
	}

	if f.ActorID != nil {
		preds = append(preds, sq.Eq{"actor_id": f.ActorID})
	}

	if f.Action != nil {
		preds = append(preds, sq.Eq{"action": f.Action})
	}

	if f.Since != nil {
		preds = append(preds, sq.GtOrEq{"timestamp": f.Since})
	}

	if f.Until != nil {
		preds = append(preds, sq.LtOrEq{"timestamp": f.Until})
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
