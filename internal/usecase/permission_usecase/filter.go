package permission_usecase

import (
	"github.com/Housiadas/cerberus/internal/core/permission"
	"github.com/Housiadas/cerberus/internal/errs"
	"github.com/Housiadas/cerberus/internal/types/name"
	"github.com/google/uuid"
)

type AppQueryParams struct {
	Cursor  string
	Limit   string
	OrderBy string
	ID      string
	Name    string
}

func parseFilter(qp AppQueryParams) (permission.QueryFilter, error) {
	var (
		fieldErrors errs.FieldErrors
		filter      permission.QueryFilter
	)

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		switch err {
		case nil:
			filter.ID = &id
		default:
			fieldErrors.Add("permission_id", err)
		}
	}

	if qp.Name != "" {
		n, err := name.Parse(qp.Name)
		switch err {
		case nil:
			filter.Name = &n
		default:
			fieldErrors.Add("name", err)
		}
	}

	if fieldErrors != nil {
		return permission.QueryFilter{}, fieldErrors.ToError()
	}

	return filter, nil
}
