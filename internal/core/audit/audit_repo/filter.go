package audit_repo

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/Housiadas/cerberus/internal/core/audit"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/order"
)

func applyCursor(cur cursor.Cursor, orderBy order.By, data map[string]any, buf *bytes.Buffer) {
	if !cur.HasCursor() {
		return
	}

	by, exists := getOrderByFields()[orderBy.Field]
	if !exists {
		return
	}

	data["cursor_value"] = cur.FieldValue()
	data["cursor_id"] = cur.ID()

	op := ">"
	if orderBy.Direction == order.DESC {
		op = "<"
	}

	fmt.Fprintf(buf, " AND (%s, id) %s (:cursor_value, :cursor_id)", by, op)
}

func applyFilter(filter audit.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	wc := []string{"1=1"}

	if filter.ObjID != nil {
		data["obj_id"] = filter.ObjID

		wc = append(wc, "obj_id = :obj_id")
	}

	if filter.ObjEntity != nil {
		data["obj_domain"] = filter.ObjEntity.String()

		wc = append(wc, "obj_domain = :obj_domain")
	}

	if filter.ObjName != nil {
		data["obj_name"] = fmt.Sprintf("%%%s%%", filter.ObjName.String())

		wc = append(wc, "obj_name LIKE :obj_name")
	}

	if filter.ActorID != nil {
		data["actor_id"] = filter.ActorID

		wc = append(wc, "actor_id = :actor_id")
	}

	if filter.Action != nil {
		data["action"] = filter.Action

		wc = append(wc, "action = :action")
	}

	if filter.Since != nil {
		data["since"] = filter.Since

		wc = append(wc, "timestamp >= :since")
	}

	if filter.Until != nil {
		data["until"] = filter.Until

		wc = append(wc, "timestamp <= :until")
	}

	buf.WriteString(" WHERE ")
	buf.WriteString(strings.Join(wc, " AND "))
}
