package user_roles_usecase

import (
	"github.com/google/uuid"

	"github.com/Housiadas/cerberus/internal/common/validation"
	ur "github.com/Housiadas/cerberus/internal/core/domain/user_roles"
)

type AppQueryParams struct {
	Page    string
	Rows    string
	OrderBy string
	UserID  string
	RoleID  string
}

func parseFilter(qp AppQueryParams) (ur.QueryFilter, error) {
	var fieldErrors validation.FieldErrors
	var filter ur.QueryFilter

	if qp.UserID != "" {
		id, err := uuid.Parse(qp.UserID)
		switch err {
		case nil:
			filter.UserID = &id
		default:
			fieldErrors.Add("user_id", err)
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

	if fieldErrors != nil {
		return ur.QueryFilter{}, fieldErrors.ToError()
	}

	return filter, nil
}
