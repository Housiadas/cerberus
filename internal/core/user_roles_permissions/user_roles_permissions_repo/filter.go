package user_roles_permissions_repo

import (
	"bytes"
	"fmt"
	"strings"

	urp "github.com/Housiadas/cerberus/internal/core/user_roles_permissions"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/order"
)

func applyCursor(cur cursor.Cursor, ob order.By, data map[string]any, buf *bytes.Buffer) {
	if !cur.HasCursor() {
		return
	}

	by, exists := getOrderFields()[ob.Field]
	if !exists {
		return
	}

	data["cursor_value"] = cur.FieldValue()
	data["cursor_id"] = cur.ID()

	op := ">"
	if ob.Direction == order.DESC {
		op = "<"
	}

	fmt.Fprintf(buf, " AND (%s, user_id) %s (:cursor_value, :cursor_id)", by, op)
}

func applyFilter(filter urp.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	wc := []string{"1=1"}

	if filter.UserID != nil {
		data["user_id"] = *filter.UserID

		wc = append(wc, "user_id = :user_id")
	}

	if filter.UserName != nil {
		data["user_name"] = fmt.Sprintf("%%%s%%", *filter.UserName)

		wc = append(wc, "user_name LIKE :user_name")
	}

	if filter.UserEmail != nil {
		data["user_email"] = filter.UserEmail.Address

		wc = append(wc, "user_email = :user_email")
	}

	if filter.RoleID != nil {
		data["role_id"] = *filter.RoleID

		wc = append(wc, "role_id = :role_id")
	}

	if filter.RoleName != nil {
		data["role_name"] = fmt.Sprintf("%%%s%%", *filter.RoleName)

		wc = append(wc, "role_name LIKE :role_name")
	}

	if filter.PermissionID != nil {
		data["permission_id"] = *filter.PermissionID

		wc = append(wc, "permission_id = :permission_id")
	}

	if filter.PermissionName != nil {
		data["permission_name"] = fmt.Sprintf("%%%s%%", *filter.PermissionName)

		wc = append(wc, "permission_name LIKE :permission_name")
	}

	buf.WriteString(" WHERE ")
	buf.WriteString(strings.Join(wc, " AND "))
}
