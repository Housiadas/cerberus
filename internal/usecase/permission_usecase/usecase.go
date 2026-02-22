// Package permission_usecase maintains the use case layer api the model
package permission_usecase

import (
	"context"

	"github.com/Housiadas/cerberus/internal/core/service/permission_service"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/web"
	"github.com/Housiadas/cerberus/pkg/web/errs"
	"github.com/google/uuid"
)

// UseCase manages the set of cli layer api functions for the permission core.
type UseCase struct {
	permissionService *permission_service.Service
}

// NewUseCase constructs a permission cli API for use.
func NewUseCase(permissionService *permission_service.Service) *UseCase {
	return &UseCase{
		permissionService: permissionService,
	}
}

// Create adds a new permission to the system.
func (uc *UseCase) Create(ctx context.Context, nperm NewPermission) (Permission, error) {
	np, err := toBusNewPermission(nperm)
	if err != nil {
		return Permission{}, errs.New(errs.InvalidArgument, err)
	}

	perm, err := uc.permissionService.Create(ctx, np)
	if err != nil {
		return Permission{}, errs.Errorf(errs.Internal, "create: perm[%+v]: %s", perm, err)
	}

	return toAppPermission(perm), nil
}

// Update updates an existing permission.
func (uc *UseCase) Update(
	ctx context.Context,
	res UpdatePermission,
	permissionID string,
) (Permission, error) {
	up, err := toBusUpdatePermission(res)
	if err != nil {
		return Permission{}, errs.New(errs.InvalidArgument, err)
	}

	permissionUUID, err := uuid.Parse(permissionID)
	if err != nil {
		return Permission{}, errs.Errorf(errs.InvalidArgument, "could not parse uuid: %s", err)
	}

	perm, err := uc.permissionService.QueryByID(ctx, permissionUUID)
	if err != nil {
		return Permission{}, errs.Errorf(errs.Internal, "permission query by id: %s", err)
	}

	updPerm, err := uc.permissionService.Update(ctx, perm, up)
	if err != nil {
		return Permission{}, errs.Errorf(
			errs.Internal,
			"update: permissionID[%s] up[%+v]: %s",
			permissionID,
			up,
			err,
		)
	}

	return toAppPermission(updPerm), nil
}

// Delete removes a permission from the system.
func (uc *UseCase) Delete(ctx context.Context, permissionID string) error {
	permissionUUID, err := uuid.Parse(permissionID)
	if err != nil {
		return errs.Errorf(errs.InvalidArgument, "could not parse uuid: %s", err)
	}

	perm, err := uc.permissionService.QueryByID(ctx, permissionUUID)
	if err != nil {
		return errs.Errorf(errs.Internal, "permission query by id: %s", err)
	}

	err = uc.permissionService.Delete(ctx, perm)
	if err != nil {
		return errs.Errorf(errs.Internal, "delete: permissionID[%s]: %s", permissionUUID, err)
	}

	return nil
}

// QueryByID returns a permission by its ID.
func (uc *UseCase) QueryByID(ctx context.Context, permissionID string) (Permission, error) {
	permissionUUID, err := uuid.Parse(permissionID)
	if err != nil {
		return Permission{}, errs.Errorf(errs.InvalidArgument, "could not parse uuid: %s", err)
	}

	perm, err := uc.permissionService.QueryByID(ctx, permissionUUID)
	if err != nil {
		return Permission{}, errs.Errorf(errs.Internal, "permission query by id: %s", err)
	}

	return toAppPermission(perm), nil
}

// Query returns a list of permissions with paging.
func (uc *UseCase) Query(ctx context.Context, qp AppQueryParams) (web.Result[Permission], error) {
	p, err := web.Parse(qp.Page, qp.Rows)
	if err != nil {
		return web.Result[Permission]{}, errs.NewFieldErrors("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return web.Result[Permission]{}, err
	}

	orderBy, err := order.Parse(getOrderByFields(), qp.OrderBy, getDefaultOrderBy())
	if err != nil {
		return web.Result[Permission]{}, errs.NewFieldErrors("order", err)
	}

	perms, err := uc.permissionService.Query(ctx, filter, orderBy, p)
	if err != nil {
		return web.Result[Permission]{}, errs.Errorf(errs.Internal, "query: %s", err)
	}

	total, err := uc.permissionService.Count(ctx, filter)
	if err != nil {
		return web.Result[Permission]{}, errs.Errorf(errs.Internal, "count: %s", err)
	}

	return web.NewResult(toAppPermissions(perms), total, p), nil
}
