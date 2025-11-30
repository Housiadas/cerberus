// Package role_usecase maintains the cli layer api for the user core.
package role_usecase

import (
	"context"

	ctxPck "github.com/Housiadas/cerberus/internal/common/context"
	"github.com/Housiadas/cerberus/internal/common/validation"
	"github.com/Housiadas/cerberus/internal/core/service/role_service"
	"github.com/Housiadas/cerberus/pkg/errs"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/page"
)

// UserCase manages the set of cli layer api functions for the user core.
type UserCase struct {
	roleService *role_service.Service
}

// NewUseCase constructs a user cli API for use.
func NewUseCase(roleService *role_service.Service) *UserCase {
	return &UserCase{
		roleService: roleService,
	}
}

// Create adds a new role to the system.
func (uc *UserCase) Create(ctx context.Context, nrole NewRole) (Role, error) {
	nc, err := toBusNewRole(nrole)
	if err != nil {
		return Role{}, errs.New(errs.InvalidArgument, err)
	}

	rol, err := uc.roleService.Create(ctx, nc)
	if err != nil {
		return Role{}, errs.Newf(errs.Internal, "create: rol[%+v]: %s", rol, err)
	}

	return toAppRole(rol), nil
}

// Update updates an existing role.
func (uc *UserCase) Update(ctx context.Context, app UpdateRole) (Role, error) {
	uu, err := toBusUpdateUser(app)
	if err != nil {
		return Role{}, errs.New(errs.InvalidArgument, err)
	}

	roleID, err := ctxPck.GetRoleID(ctx)
	if err != nil {
		return Role{}, errs.Newf(errs.Internal, "roleID not in ctx: %s", err)
	}

	role, err := uc.roleService.QueryByID(ctx, roleID)
	if err != nil {
		return Role{}, errs.Newf(errs.Internal, "role query by id: %s", err)
	}

	updRole, err := uc.roleService.Update(ctx, role, uu)
	if err != nil {
		return Role{}, errs.Newf(errs.Internal, "update: userID[%s] uu[%+v]: %s", roleID, uu, err)
	}

	return toAppRole(updRole), nil
}

// Delete removes a role from the system.
func (uc *UserCase) Delete(ctx context.Context) error {
	rl, err := ctxPck.GetRole(ctx)
	if err != nil {
		return errs.Newf(errs.Internal, "userID missing in context: %s", err)
	}

	if err := uc.roleService.Delete(ctx, rl); err != nil {
		return errs.Newf(errs.Internal, "delete: roleID[%s]: %s", rl.ID, err)
	}

	return nil
}

// QueryByID returns a role by its ID
func (uc *UserCase) QueryByID(ctx context.Context) (Role, error) {
	roleID, err := ctxPck.GetRoleID(ctx)
	if err != nil {
		return Role{}, errs.Newf(errs.Internal, "roleID not in ctx: %s", err)
	}

	role, err := uc.roleService.QueryByID(ctx, roleID)
	if err != nil {
		return Role{}, errs.Newf(errs.Internal, "role query by id: %s", err)
	}

	return toAppRole(role), nil
}

// Query returns a list of roles with paging.
func (uc *UserCase) Query(ctx context.Context, qp AppQueryParams) (page.Result[Role], error) {
	p, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return page.Result[Role]{}, validation.NewFieldErrors("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return page.Result[Role]{}, err
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return page.Result[Role]{}, validation.NewFieldErrors("order", err)
	}

	usrs, err := uc.roleService.Query(ctx, filter, orderBy, p)
	if err != nil {
		return page.Result[Role]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := uc.roleService.Count(ctx, filter)
	if err != nil {
		return page.Result[Role]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return page.NewResult(toAppRoles(usrs), total, p), nil
}
