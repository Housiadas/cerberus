package user_roles_permissions_usecase

import (
	"net/mail"

	"github.com/google/uuid"

	"github.com/Housiadas/cerberus/internal/common/validation"
	"github.com/Housiadas/cerberus/internal/core/domain/name"
	urp "github.com/Housiadas/cerberus/internal/core/domain/user_roles_permissions"
)

type AppQueryParams struct {
	Page           string
	Rows           string
	OrderBy        string
	UserID         string
	UserName       string
	UserEmail      string
	RoleID         string
	RoleName       string
	PermissionID   string
	PermissionName string
}

func parseFilter(qp AppQueryParams) (urp.QueryFilter, error) {
	var fieldErrors validation.FieldErrors
	var filter urp.QueryFilter

	if qp.UserID != "" {
		id, err := uuid.Parse(qp.UserID)
		switch err {
		case nil:
			filter.UserID = &id
		default:
			fieldErrors.Add("user_id", err)
		}
	}

	if qp.UserName != "" {
		n, err := name.Parse(qp.UserName)
		switch err {
		case nil:
			filter.UserName = &n
		default:
			fieldErrors.Add("user_name", err)
		}
	}

	if qp.UserEmail != "" {
		addr, err := mail.ParseAddress(qp.UserEmail)
		switch err {
		case nil:
			filter.UserEmail = addr
		default:
			fieldErrors.Add("user_email", err)
		}
	}

	if qp.RoleID != "" {
		id, err := uuid.Parse(qp.RoleID)
		switch err {
		case nil:
			filter.RoleID = &id
		default:
			fieldErrors.Add("role_id", err)
		}
	}

	if qp.RoleName != "" {
		n, err := name.Parse(qp.RoleName)
		switch err {
		case nil:
			filter.RoleName = &n
		default:
			fieldErrors.Add("role_name", err)
		}
	}

	if qp.PermissionID != "" {
		id, err := uuid.Parse(qp.PermissionID)
		switch err {
		case nil:
			filter.PermissionID = &id
		default:
			fieldErrors.Add("permission_id", err)
		}
	}

	if qp.PermissionName != "" {
		n, err := name.Parse(qp.PermissionName)
		switch err {
		case nil:
			filter.PermissionName = &n
		default:
			fieldErrors.Add("permission_name", err)
		}
	}

	if fieldErrors != nil {
		return urp.QueryFilter{}, fieldErrors.ToError()
	}

	return filter, nil
}
