// Package user_roles_permissions_usecase maintains the usecase layer api for the view
package user_roles_permissions_usecase

import (
	"context"

	"github.com/Housiadas/cerberus/internal/core/service/role_permissions_service"
	"github.com/Housiadas/cerberus/internal/core/service/user_roles_permissions_service"
	"github.com/Housiadas/cerberus/internal/core/service/user_roles_service"
	"github.com/Housiadas/cerberus/internal/utils/errs"
	"github.com/Housiadas/cerberus/internal/utils/page"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/google/uuid"
)

// UseCase composes user_roles_permissions, user_roles and role_permissions services.
type UseCase struct {
	service          *user_roles_permissions_service.Service
	userRolesService *user_roles_service.Service
	rolePermsService *role_permissions_service.Service
}

// Config holds the dependencies for the use case.
type Config struct {
	Service          *user_roles_permissions_service.Service
	UserRolesService *user_roles_service.Service
	RolePermsService *role_permissions_service.Service
}

func NewUseCase(cfg Config) *UseCase {
	return &UseCase{
		service:          cfg.Service,
		userRolesService: cfg.UserRolesService,
		rolePermsService: cfg.RolePermsService,
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

// QueryPermissionsByUserID returns all permission names for the given user.
func (uc *UseCase) QueryPermissionsByUserID(ctx context.Context, userID string) ([]string, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errs.Errorf(errs.InvalidArgument, "could not parse uuid: %s", err)
	}

	permissions, err := uc.service.QueryPermissionsByUserID(ctx, userUUID)
	if err != nil {
		return nil, errs.Errorf(errs.Internal, "query_permissions_by_user_id: %s", err)
	}

	return permissions, nil
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

// AddUserRole assigns a role to a user.
func (uc *UseCase) AddUserRole(ctx context.Context, userID, roleID string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return errs.Errorf(errs.InvalidArgument, "could not parse user uuid: %s", err)
	}

	roleUUID, err := uuid.Parse(roleID)
	if err != nil {
		return errs.Errorf(errs.InvalidArgument, "could not parse role uuid: %s", err)
	}

	err = uc.userRolesService.Add(ctx, userUUID, roleUUID)
	if err != nil {
		return errs.Errorf(errs.Internal, "add_user_role: %s", err)
	}

	return nil
}

// RemoveUserRole removes a role from a user.
func (uc *UseCase) RemoveUserRole(ctx context.Context, userID, roleID string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return errs.Errorf(errs.InvalidArgument, "could not parse user uuid: %s", err)
	}

	roleUUID, err := uuid.Parse(roleID)
	if err != nil {
		return errs.Errorf(errs.InvalidArgument, "could not parse role uuid: %s", err)
	}

	err = uc.userRolesService.Remove(ctx, userUUID, roleUUID)
	if err != nil {
		return errs.Errorf(errs.Internal, "remove_user_role: %s", err)
	}

	return nil
}

// AddRolePermission assigns a permission to a role.
func (uc *UseCase) AddRolePermission(ctx context.Context, roleID, permissionID string) error {
	roleUUID, err := uuid.Parse(roleID)
	if err != nil {
		return errs.Errorf(errs.InvalidArgument, "could not parse role uuid: %s", err)
	}

	permUUID, err := uuid.Parse(permissionID)
	if err != nil {
		return errs.Errorf(errs.InvalidArgument, "could not parse permission uuid: %s", err)
	}

	err = uc.rolePermsService.Add(ctx, roleUUID, permUUID)
	if err != nil {
		return errs.Errorf(errs.Internal, "add_role_permission: %s", err)
	}

	return nil
}

// RemoveRolePermission removes a permission from a role.
func (uc *UseCase) RemoveRolePermission(ctx context.Context, roleID, permissionID string) error {
	roleUUID, err := uuid.Parse(roleID)
	if err != nil {
		return errs.Errorf(errs.InvalidArgument, "could not parse role uuid: %s", err)
	}

	permUUID, err := uuid.Parse(permissionID)
	if err != nil {
		return errs.Errorf(errs.InvalidArgument, "could not parse permission uuid: %s", err)
	}

	err = uc.rolePermsService.Remove(ctx, roleUUID, permUUID)
	if err != nil {
		return errs.Errorf(errs.Internal, "remove_role_permission: %s", err)
	}

	return nil
}
