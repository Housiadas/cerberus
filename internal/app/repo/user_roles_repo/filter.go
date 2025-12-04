package user_roles_repo

import (
	"bytes"
	"strings"

	ur "github.com/Housiadas/cerberus/internal/core/domain/user_roles"
)

func applyFilter(filter ur.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.UserID != nil {
		data["user_id"] = *filter.UserID
		wc = append(wc, "user_id = :user_id")
	}
	if filter.RoleID != nil {
		data["role_id"] = *filter.RoleID
		wc = append(wc, "role_id = :role_id")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
