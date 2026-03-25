package account_usecase

import (
	"github.com/Housiadas/cerberus/internal/core/account"
	"github.com/Housiadas/cerberus/internal/errs"
	"github.com/google/uuid"
)

// AppQueryParams holds query parameters from the API layer.
type AppQueryParams struct {
	Cursor  string
	Limit   string
	OrderBy string
	ID      string
	Name    string
}

func parseFilter(qp AppQueryParams) (account.QueryFilter, error) {
	var (
		fieldErrors errs.FieldErrors
		filter      account.QueryFilter
	)

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		switch err {
		case nil:
			filter.ID = &id
		default:
			fieldErrors.Add("account_id", err)
		}
	}

	if qp.Name != "" {
		filter.Name = &qp.Name
	}

	if fieldErrors != nil {
		return account.QueryFilter{}, fieldErrors.ToError()
	}

	return filter, nil
}
