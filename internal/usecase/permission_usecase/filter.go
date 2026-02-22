package permission_usecase

import (
	"github.com/Housiadas/cerberus/internal/core/domain/name"
	"github.com/Housiadas/cerberus/internal/core/domain/permission"
	"github.com/Housiadas/cerberus/pkg/web/errs"
	"github.com/google/uuid"
)

type AppQueryParams struct {
	Page    string
	Rows    string
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
