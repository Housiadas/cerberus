// Package role_usecase maintains the usecase layer api the model
package role_usecase

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/audit"
	"github.com/Housiadas/cerberus/internal/core/domain/role"
	"github.com/Housiadas/cerberus/internal/core/service/role_service"
	errs2 "github.com/Housiadas/cerberus/internal/errs"
	"github.com/Housiadas/cerberus/internal/eventbus"
	"github.com/Housiadas/cerberus/internal/types/entity"
	"github.com/Housiadas/cerberus/internal/types/event"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
)

type UseCase struct {
	log         logger.Logger
	tx          pgsql.Beginner
	roleService *role_service.Service
	dispatcher  *eventbus.EventDispatcher
}

func NewUseCase(
	log logger.Logger,
	roleService *role_service.Service,
	dispatcher *eventbus.EventDispatcher,
	tx pgsql.Beginner,
) *UseCase {
	return &UseCase{
		log:         log,
		tx:          tx,
		roleService: roleService,
		dispatcher:  dispatcher,
	}
}

// Create adds a new role to the system.
func (uc *UseCase) Create(ctx context.Context, nrole NewRole) (Role, error) {
	nc, err := toBusNewRole(nrole)
	if err != nil {
		return Role{}, errs2.New(errs2.InvalidArgument, errs2.CodeValidation, err)
	}

	var rol role.Role

	txErr := pgsql.RunInTx(ctx, uc.log, uc.tx, func(tran pgsql.CommitRollbacker) error {
		roleServiceTx, initErr := uc.roleService.NewWithTx(tran)
		if initErr != nil {
			return errs2.Errorf(errs2.Internal, errs2.CodeInternal, "role tx: %s", initErr)
		}

		rol, err = roleServiceTx.Create(ctx, nc)
		if err != nil {
			return errs2.Errorf(
				errs2.Internal,
				errs2.CodeInternal,
				"create: rol[%+v]: %s",
				rol,
				err,
			)
		}

		return uc.dispatcher.Dispatch(ctx, tran, newRoleEvent(rol, audit.ActionCreate))
	})
	if txErr != nil {
		return Role{}, fmt.Errorf("create role: %w", txErr)
	}

	return toAppRole(rol), nil
}

// Update updates an existing role.
func (uc *UseCase) Update(ctx context.Context, res UpdateRole, roleID string) (Role, error) {
	uu, err := toBusUpdateUser(res)
	if err != nil {
		return Role{}, errs2.New(errs2.InvalidArgument, errs2.CodeValidation, err)
	}

	roleUUID, err := uuid.Parse(roleID)
	if err != nil {
		return Role{}, errs2.Errorf(
			errs2.InvalidArgument,
			errs2.CodeValidation,
			"could not parse uuid: %s",
			err,
		)
	}

	currentRole, err := uc.roleService.QueryByID(ctx, roleUUID)
	if err != nil {
		return Role{}, errs2.Errorf(errs2.Internal, errs2.CodeInternal, "role query by id: %s", err)
	}

	var updRole role.Role

	txErr := pgsql.RunInTx(ctx, uc.log, uc.tx, func(tran pgsql.CommitRollbacker) error {
		roleServiceTx, initErr := uc.roleService.NewWithTx(tran)
		if initErr != nil {
			return errs2.Errorf(errs2.Internal, errs2.CodeInternal, "role tx: %s", initErr)
		}

		updRole, err = roleServiceTx.Update(ctx, currentRole, uu)
		if err != nil {
			return errs2.Errorf(
				errs2.Internal,
				errs2.CodeInternal,
				"update: roleID[%s] uu[%+v]: %s",
				roleID,
				uu,
				err,
			)
		}

		return uc.dispatcher.Dispatch(ctx, tran, newRoleEvent(updRole, audit.ActionUpdate))
	})
	if txErr != nil {
		return Role{}, fmt.Errorf("update role: %w", txErr)
	}

	return toAppRole(updRole), nil
}

// Delete removes a role from the system.
func (uc *UseCase) Delete(ctx context.Context, roleID string) error {
	roleUUID, err := uuid.Parse(roleID)
	if err != nil {
		return errs2.Errorf(
			errs2.InvalidArgument,
			errs2.CodeValidation,
			"could not parse uuid: %s",
			err,
		)
	}

	currentRole, err := uc.roleService.QueryByID(ctx, roleUUID)
	if err != nil {
		return errs2.Errorf(errs2.Internal, errs2.CodeInternal, "role query by id: %s", err)
	}

	txErr := pgsql.RunInTx(ctx, uc.log, uc.tx, func(tran pgsql.CommitRollbacker) error {
		roleServiceTx, initErr := uc.roleService.NewWithTx(tran)
		if initErr != nil {
			return errs2.Errorf(errs2.Internal, errs2.CodeInternal, "role tx: %s", initErr)
		}

		deleteErr := roleServiceTx.Delete(ctx, currentRole)
		if deleteErr != nil {
			return errs2.Errorf(
				errs2.Internal,
				errs2.CodeInternal,
				"delete: roleID[%s]: %s",
				roleUUID,
				deleteErr,
			)
		}

		return uc.dispatcher.Dispatch(ctx, tran, newRoleEvent(currentRole, audit.ActionDelete))
	})
	if txErr != nil {
		return fmt.Errorf("delete role: %w", txErr)
	}

	return nil
}

// QueryByID returns a role by its ID.
func (uc *UseCase) QueryByID(ctx context.Context, roleID string) (Role, error) {
	roleUUID, err := uuid.Parse(roleID)
	if err != nil {
		return Role{}, errs2.Errorf(
			errs2.InvalidArgument,
			errs2.CodeValidation,
			"could not parse uuid: %s",
			err,
		)
	}

	rol, err := uc.roleService.QueryByID(ctx, roleUUID)
	if err != nil {
		return Role{}, errs2.Errorf(errs2.Internal, errs2.CodeInternal, "role query by id: %s", err)
	}

	return toAppRole(rol), nil
}

// Query returns a list of roles with cursor-based paging.
func (uc *UseCase) Query(ctx context.Context, qp AppQueryParams) (cursor.Result[Role], error) {
	cur, err := cursor.Parse(qp.Cursor, qp.Limit)
	if err != nil {
		return cursor.Result[Role]{}, errs2.NewFieldErrors("cursor", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return cursor.Result[Role]{}, err
	}

	orderBy, err := order.Parse(getOrderByFields(), qp.OrderBy, getDefaultOrderBy())
	if err != nil {
		return cursor.Result[Role]{}, errs2.NewFieldErrors("order", err)
	}

	roles, err := uc.roleService.Query(ctx, filter, orderBy, cur)
	if err != nil {
		return cursor.Result[Role]{}, errs2.Errorf(
			errs2.Internal,
			errs2.CodeInternal,
			"query: %s",
			err,
		)
	}

	return cursor.NewResult(
		toAppRoles(roles),
		cur.Limit(),
		cur,
		orderBy,
		func(r Role) string { return r.ID },
		func(r Role) any { return r.ID },
	), nil
}

func newRoleEvent(rl role.Role, action string) event.DomainEvent {
	return event.DomainEvent{
		AggregateID: rl.ID(),
		Payload:     rl,
		ObjEntity:   entity.New(entity.RoleEntity),
		ObjName:     rl.Name(),
		Action:      action,
		Message:     "role " + action,
	}
}
