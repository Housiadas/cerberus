package user_roles_permissions_repo

import (
	"bytes"
	"fmt"
	"strings"

	urp "github.com/Housiadas/cerberus/internal/core/domain/user_roles_permissions"
)

func applyFilter(filter urp.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

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

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
