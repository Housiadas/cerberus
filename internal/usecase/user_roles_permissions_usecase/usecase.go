// Package user_roles_permissions_usecase maintains the usecase layer api for the view
package user_roles_permissions_usecase

import (
	"context"

	"github.com/Housiadas/cerberus/internal/core/service/user_roles_permissions_service"
	"github.com/Housiadas/cerberus/internal/utils/errs"
	"github.com/Housiadas/cerberus/internal/utils/page"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/google/uuid"
)

// UseCase provides read access to the user-roles-permissions view.
type UseCase struct {
	service *user_roles_permissions_service.Service
}

// Config holds the dependencies for the use case.
type Config struct {
	Service *user_roles_permissions_service.Service
}

func NewUseCase(cfg Config) *UseCase {
	return &UseCase{
		service: cfg.Service,
	}
}

// Query returns a list of rows with paging.
func (uc *UseCase) Query(
	ctx context.Context,
	qp AppQueryParams,
) (page.Result[UserRolesPermissions], error) {
	p, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return page.Result[UserRolesPermissions]{}, errs.NewFieldErrors("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return page.Result[UserRolesPermissions]{}, err
	}

	ob, err := order.Parse(getOrderByFields(), qp.OrderBy, getDefaultOrderBy())
	if err != nil {
		return page.Result[UserRolesPermissions]{}, errs.NewFieldErrors("order", err)
	}

	rows, err := uc.service.Query(ctx, filter, ob, p)
	if err != nil {
		return page.Result[UserRolesPermissions]{}, errs.Errorf(errs.Internal, "query: %s", err)
	}

	total, err := uc.service.Count(ctx, filter)
	if err != nil {
		return page.Result[UserRolesPermissions]{}, errs.Errorf(errs.Internal, "count: %s", err)
	}

	return page.NewResult(toManyUserRolesPermissions(rows), total, p), nil
}

// QueryPermissionsByUserID returns all permissions (id and name) for the given user.
func (uc *UseCase) QueryPermissionsByUserID(
	ctx context.Context,
	userID string,
) ([]Permission, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errs.Errorf(errs.InvalidArgument, "could not parse uuid: %s", err)
	}

	perms, err := uc.service.QueryPermissionsByUserID(ctx, userUUID)
	if err != nil {
		return nil, errs.Errorf(errs.Internal, "query_permissions_by_user_id: %s", err)
	}

	result := make([]Permission, len(perms))
	for i, p := range perms {
		result[i] = Permission{
			ID:   p.ID.String(),
			Name: p.Name,
		}
	}

	return result, nil
}

// HasPermission checks if the user has the specified permission.
func (uc *UseCase) HasPermission(ctx context.Context, userID, permissionName string) (bool, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return false, errs.Errorf(errs.InvalidArgument, "could not parse uuid: %s", err)
	}

	hasPermission, err := uc.service.HasPermission(ctx, userUUID, permissionName)
	if err != nil {
		return false, errs.Errorf(errs.Internal, "has_permission: %s", err)
	}

	return hasPermission, nil
}
