// Package permission_usecase maintains the use case layer api the model
package permission_usecase

import (
	"context"

	ctxPck "github.com/Housiadas/cerberus/internal/common/context"
	"github.com/Housiadas/cerberus/internal/common/validation"
	"github.com/Housiadas/cerberus/internal/core/service/permission_service"
	"github.com/Housiadas/cerberus/pkg/errs"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/page"
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
		return Permission{}, errs.Newf(errs.Internal, "create: perm[%+v]: %s", perm, err)
	}

	return toAppPermission(perm), nil
}

// Update updates an existing permission.
func (uc *UseCase) Update(ctx context.Context, app UpdatePermission) (Permission, error) {
	up, err := toBusUpdatePermission(app)
	if err != nil {
		return Permission{}, errs.New(errs.InvalidArgument, err)
	}

	permissionID, err := ctxPck.GetPermissionID(ctx)
	if err != nil {
		return Permission{}, errs.Newf(errs.Internal, "permissionID not in ctx: %s", err)
	}

	perm, err := uc.permissionService.QueryByID(ctx, permissionID)
	if err != nil {
		return Permission{}, errs.Newf(errs.Internal, "permission query by id: %s", err)
	}

	updPerm, err := uc.permissionService.Update(ctx, perm, up)
	if err != nil {
		return Permission{}, errs.Newf(errs.Internal, "update: permissionID[%s] up[%+v]: %s", permissionID, up, err)
	}

	return toAppPermission(updPerm), nil
}

// Delete removes a permission from the system.
func (uc *UseCase) Delete(ctx context.Context) error {
	p, err := ctxPck.GetPermission(ctx)
	if err != nil {
		return errs.Newf(errs.Internal, "permission missing in context: %s", err)
	}

	if err := uc.permissionService.Delete(ctx, p); err != nil {
		return errs.Newf(errs.Internal, "delete: permissionID[%s]: %s", p.ID, err)
	}

	return nil
}

// QueryByID returns a permission by its ID
func (uc *UseCase) QueryByID(ctx context.Context) (Permission, error) {
	permissionID, err := ctxPck.GetPermissionID(ctx)
	if err != nil {
		return Permission{}, errs.Newf(errs.Internal, "permissionID not in ctx: %s", err)
	}

	perm, err := uc.permissionService.QueryByID(ctx, permissionID)
	if err != nil {
		return Permission{}, errs.Newf(errs.Internal, "permission query by id: %s", err)
	}

	return toAppPermission(perm), nil
}

// Query returns a list of permissions with paging.
func (uc *UseCase) Query(ctx context.Context, qp AppQueryParams) (page.Result[Permission], error) {
	p, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return page.Result[Permission]{}, validation.NewFieldErrors("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return page.Result[Permission]{}, err
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return page.Result[Permission]{}, validation.NewFieldErrors("order", err)
	}

	perms, err := uc.permissionService.Query(ctx, filter, orderBy, p)
	if err != nil {
		return page.Result[Permission]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := uc.permissionService.Count(ctx, filter)
	if err != nil {
		return page.Result[Permission]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return page.NewResult(toAppPermissions(perms), total, p), nil
}
