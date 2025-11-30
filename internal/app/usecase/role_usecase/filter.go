package user_usecase

import (
	"github.com/google/uuid"

	"github.com/Housiadas/cerberus/internal/common/validation"
	"github.com/Housiadas/cerberus/internal/core/domain/name"
	"github.com/Housiadas/cerberus/internal/core/domain/user"
)

type AppQueryParams struct {
	Page    string
	Rows    string
	OrderBy string
	ID      string
	Name    string
}

func parseFilter(qp AppQueryParams) (user.QueryFilter, error) {
	var fieldErrors validation.FieldErrors
	var filter user.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		switch err {
		case nil:
			filter.ID = &id
		default:
			fieldErrors.Add("user_id", err)
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
		return user.QueryFilter{}, fieldErrors.ToError()
	}

	return filter, nil
}
