// Package user_roles_usecase maintains the use case layer api for the view
package user_roles_usecase

import (
	"context"

	"github.com/Housiadas/cerberus/internal/common/validation"
	"github.com/Housiadas/cerberus/internal/core/service/user_roles_service"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/web"
	"github.com/Housiadas/cerberus/pkg/web/errs"
	"github.com/google/uuid"
)

// UseCase manages the set of cli layer api functions for the view.
type UseCase struct {
	service *user_roles_service.Service
}

// NewUseCase constructs the API for use.
func NewUseCase(service *user_roles_service.Service) *UseCase {
	return &UseCase{
		service: service,
	}
}

// Query returns a list of rows with paging.
func (uc *UseCase) Query(ctx context.Context, qp AppQueryParams) (web.Result[UserRole], error) {
	p, err := web.Parse(qp.Page, qp.Rows)
	if err != nil {
		return web.Result[UserRole]{}, validation.ErrorfieldErrors("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return web.Result[UserRole]{}, err
	}

	ob, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return web.Result[UserRole]{}, validation.ErrorfieldErrors("order", err)
	}

	rows, err := uc.service.Query(ctx, filter, ob, p)
	if err != nil {
		return web.Result[UserRole]{}, errs.Errorf(errs.Internal, "query: %s", err)
	}

	total, err := uc.service.Count(ctx, filter)
	if err != nil {
		return web.Result[UserRole]{}, errs.Errorf(errs.Internal, "count: %s", err)
	}

	return web.NewResult(toManyUserRolesPermissions(rows), total, p), nil
}

func (uc *UseCase) GetUserRolesNames(ctx context.Context, userID string) ([]string, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errs.Errorf(errs.InvalidArgument, "could not parse uuid: %s", err)
	}

	roles, err := uc.service.GetUserRoleNames(ctx, userUUID)
	if err != nil {
		return nil, errs.Errorf(errs.Internal, "get_user_roles_names: %s", err)
	}
	return roles, nil
}
