// Package user_roles_permissions_usecase maintains the use case layer api for the view
package user_roles_permissions_usecase

import (
	"context"

	"github.com/Housiadas/cerberus/internal/common/validation"
	"github.com/Housiadas/cerberus/internal/core/service/user_roles_permissions_service"
	"github.com/Housiadas/cerberus/pkg/errs"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/page"
)

// UseCase manages the set of cli layer api functions for the view.
type UseCase struct {
	service *user_roles_permissions_service.Service
}

// NewUseCase constructs the API for use.
func NewUseCase(service *user_roles_permissions_service.Service) *UseCase {
	return &UseCase{
		service: service,
	}
}

// Query returns a list of rows with paging.
func (uc *UseCase) Query(ctx context.Context, qp AppQueryParams) (page.Result[UserRolesPermissions], error) {
	p, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return page.Result[UserRolesPermissions]{}, validation.NewFieldErrors("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return page.Result[UserRolesPermissions]{}, err
	}

	ob, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return page.Result[UserRolesPermissions]{}, validation.NewFieldErrors("order", err)
	}

	rows, err := uc.service.Query(ctx, filter, ob, p)
	if err != nil {
		return page.Result[UserRolesPermissions]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := uc.service.Count(ctx, filter)
	if err != nil {
		return page.Result[UserRolesPermissions]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return page.NewResult(toManyUserRolesPermissions(rows), total, p), nil
}
