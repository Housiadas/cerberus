// Package role_usecase maintains the use case layer api the model
package role_usecase

import (
	"context"

	"github.com/Housiadas/cerberus/pkg/web"
	"github.com/Housiadas/cerberus/pkg/web/errs"
	"github.com/google/uuid"

	"github.com/Housiadas/cerberus/internal/common/validation"
	"github.com/Housiadas/cerberus/internal/core/service/role_service"
	"github.com/Housiadas/cerberus/pkg/order"
)

// UseCase manages the set of cli layer api functions for the user core.
type UseCase struct {
	roleService *role_service.Service
}

// NewUseCase constructs a user cli API for use.
func NewUseCase(roleService *role_service.Service) *UseCase {
	return &UseCase{
		roleService: roleService,
	}
}

// Create adds a new role to the system.
func (uc *UseCase) Create(ctx context.Context, nrole NewRole) (Role, error) {
	nc, err := toBusNewRole(nrole)
	if err != nil {
		return Role{}, errs.New(errs.InvalidArgument, err)
	}

	rol, err := uc.roleService.Create(ctx, nc)
	if err != nil {
		return Role{}, errs.Errorf(errs.Internal, "create: rol[%+v]: %s", rol, err)
	}

	return toAppRole(rol), nil
}

// Update updates an existing role.
func (uc *UseCase) Update(ctx context.Context, res UpdateRole, roleID string) (Role, error) {
	uu, err := toBusUpdateUser(res)
	if err != nil {
		return Role{}, errs.New(errs.InvalidArgument, err)
	}

	roleUUID, err := uuid.Parse(roleID)
	if err != nil {
		return Role{}, errs.Errorf(errs.InvalidArgument, "could not parse uuid: %s", err)
	}

	role, err := uc.roleService.QueryByID(ctx, roleUUID)
	if err != nil {
		return Role{}, errs.Errorf(errs.Internal, "role query by id: %s", err)
	}

	updRole, err := uc.roleService.Update(ctx, role, uu)
	if err != nil {
		return Role{}, errs.Errorf(errs.Internal, "update: userID[%s] uu[%+v]: %s", roleID, uu, err)
	}

	return toAppRole(updRole), nil
}

// Delete removes a role from the system.
func (uc *UseCase) Delete(ctx context.Context, roleID string) error {
	roleUUID, err := uuid.Parse(roleID)
	if err != nil {
		return errs.Errorf(errs.InvalidArgument, "could not parse uuid: %s", err)
	}

	rl, err := uc.roleService.QueryByID(ctx, roleUUID)
	if err != nil {
		return errs.Errorf(errs.Internal, "role query by id: %s", err)
	}

	if err := uc.roleService.Delete(ctx, rl); err != nil {
		return errs.Errorf(errs.Internal, "delete: roleID[%s]: %s", rl.ID, err)
	}

	return nil
}

// QueryByID returns a role by its ID
func (uc *UseCase) QueryByID(ctx context.Context, roleID string) (Role, error) {
	roleUUID, err := uuid.Parse(roleID)
	if err != nil {
		return Role{}, errs.Errorf(errs.InvalidArgument, "could not parse uuid: %s", err)
	}

	role, err := uc.roleService.QueryByID(ctx, roleUUID)
	if err != nil {
		return Role{}, errs.Errorf(errs.Internal, "role query by id: %s", err)
	}

	return toAppRole(role), nil
}

// Query returns a list of roles with paging.
func (uc *UseCase) Query(ctx context.Context, qp AppQueryParams) (web.Result[Role], error) {
	p, err := web.Parse(qp.Page, qp.Rows)
	if err != nil {
		return web.Result[Role]{}, validation.ErrorfieldErrors("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return web.Result[Role]{}, err
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return web.Result[Role]{}, validation.ErrorfieldErrors("order", err)
	}

	usrs, err := uc.roleService.Query(ctx, filter, orderBy, p)
	if err != nil {
		return web.Result[Role]{}, errs.Errorf(errs.Internal, "query: %s", err)
	}

	total, err := uc.roleService.Count(ctx, filter)
	if err != nil {
		return web.Result[Role]{}, errs.Errorf(errs.Internal, "count: %s", err)
	}

	return web.NewResult(toAppRoles(usrs), total, p), nil
}
