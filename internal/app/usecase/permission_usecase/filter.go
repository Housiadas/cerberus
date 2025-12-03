package permission_usecase

import (
	"github.com/google/uuid"

	"github.com/Housiadas/cerberus/internal/common/validation"
	"github.com/Housiadas/cerberus/internal/core/domain/name"
	"github.com/Housiadas/cerberus/internal/core/domain/permission"
)

type AppQueryParams struct {
	Page    string
	Rows    string
	OrderBy string
	ID      string
	Name    string
}

func parseFilter(qp AppQueryParams) (permission.QueryFilter, error) {
	var fieldErrors validation.FieldErrors
	var filter permission.QueryFilter

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
