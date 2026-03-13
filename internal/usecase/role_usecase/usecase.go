// Package role_usecase maintains the usecase layer api the model
package role_usecase

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/audit"
	"github.com/Housiadas/cerberus/internal/core/domain/entity"
	"github.com/Housiadas/cerberus/internal/core/domain/role"
	"github.com/Housiadas/cerberus/internal/core/service/audit_service"
	"github.com/Housiadas/cerberus/internal/core/service/role_service"
	ctxPck "github.com/Housiadas/cerberus/internal/utils/context"
	"github.com/Housiadas/cerberus/internal/utils/errs"
	"github.com/Housiadas/cerberus/internal/utils/page"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
)

type UseCase struct {
	log         logger.Logger
	roleService *role_service.Service
	auditSvc    *audit_service.Service
	tx          pgsql.Beginner
}

func NewUseCase(
	log logger.Logger,
	roleService *role_service.Service,
	auditSvc *audit_service.Service,
	tx pgsql.Beginner,
) *UseCase {
	return &UseCase{
		log:         log,
		roleService: roleService,
		auditSvc:    auditSvc,
		tx:          tx,
	}
}

// Create adds a new role to the system.
func (uc *UseCase) Create(ctx context.Context, nrole NewRole) (Role, error) {
	nc, err := toBusNewRole(nrole)
	if err != nil {
		return Role{}, errs.New(errs.InvalidArgument, errs.CodeValidation, err)
	}

	var rol role.Role

	txErr := pgsql.RunInTx(ctx, uc.log, uc.tx, func(tran pgsql.CommitRollbacker) error {
		roleServiceTx, initErr := uc.roleService.NewWithTx(tran)
		if initErr != nil {
			return errs.Errorf(errs.Internal, errs.CodeInternal, "role tx: %s", initErr)
		}

		auditSvcTx, initErr := uc.auditSvc.NewWithTx(tran)
		if initErr != nil {
			return errs.Errorf(errs.Internal, errs.CodeInternal, "audit tx: %s", initErr)
		}

		rol, err = roleServiceTx.Create(ctx, nc)
		if err != nil {
			return errs.Errorf(
				errs.Internal,
				errs.CodeInternal,
				"create: rol[%+v]: %s",
				rol,
				err,
			)
		}

		_, auditErr := auditSvcTx.Create(ctx, uc.newRoleAudit(ctx, rol, audit.ActionCreate))
		if auditErr != nil {
			return errs.Errorf(errs.Internal, errs.CodeInternal, "audit create: %s", auditErr)
		}

		return nil
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
		return Role{}, errs.New(errs.InvalidArgument, errs.CodeValidation, err)
	}

	roleUUID, err := uuid.Parse(roleID)
	if err != nil {
		return Role{}, errs.Errorf(
			errs.InvalidArgument,
			errs.CodeValidation,
			"could not parse uuid: %s",
			err,
		)
	}

	currentRole, err := uc.roleService.QueryByID(ctx, roleUUID)
	if err != nil {
		return Role{}, errs.Errorf(errs.Internal, errs.CodeInternal, "role query by id: %s", err)
	}

	var updRole role.Role

	txErr := pgsql.RunInTx(ctx, uc.log, uc.tx, func(tran pgsql.CommitRollbacker) error {
		roleServiceTx, initErr := uc.roleService.NewWithTx(tran)
		if initErr != nil {
			return errs.Errorf(errs.Internal, errs.CodeInternal, "role tx: %s", initErr)
		}

		auditSvcTx, initErr := uc.auditSvc.NewWithTx(tran)
		if initErr != nil {
			return errs.Errorf(errs.Internal, errs.CodeInternal, "audit tx: %s", initErr)
		}

		updRole, err = roleServiceTx.Update(ctx, currentRole, uu)
		if err != nil {
			return errs.Errorf(
				errs.Internal,
				errs.CodeInternal,
				"update: roleID[%s] uu[%+v]: %s",
				roleID,
				uu,
				err,
			)
		}

		_, auditErr := auditSvcTx.Create(ctx, uc.newRoleAudit(ctx, updRole, audit.ActionUpdate))
		if auditErr != nil {
			return errs.Errorf(errs.Internal, errs.CodeInternal, "audit update: %s", auditErr)
		}

		return nil
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
		return errs.Errorf(
			errs.InvalidArgument,
			errs.CodeValidation,
			"could not parse uuid: %s",
			err,
		)
	}

	currentRole, err := uc.roleService.QueryByID(ctx, roleUUID)
	if err != nil {
		return errs.Errorf(errs.Internal, errs.CodeInternal, "role query by id: %s", err)
	}

	txErr := pgsql.RunInTx(ctx, uc.log, uc.tx, func(tran pgsql.CommitRollbacker) error {
		roleServiceTx, initErr := uc.roleService.NewWithTx(tran)
		if initErr != nil {
			return errs.Errorf(errs.Internal, errs.CodeInternal, "role tx: %s", initErr)
		}

		auditSvcTx, initErr := uc.auditSvc.NewWithTx(tran)
		if initErr != nil {
			return errs.Errorf(errs.Internal, errs.CodeInternal, "audit tx: %s", initErr)
		}

		deleteErr := roleServiceTx.Delete(ctx, currentRole)
		if deleteErr != nil {
			return errs.Errorf(
				errs.Internal,
				errs.CodeInternal,
				"delete: roleID[%s]: %s",
				roleUUID,
				deleteErr,
			)
		}

		_, auditErr := auditSvcTx.Create(ctx, uc.newRoleAudit(ctx, currentRole, audit.ActionDelete))
		if auditErr != nil {
			return errs.Errorf(errs.Internal, errs.CodeInternal, "audit delete: %s", auditErr)
		}

		return nil
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
		return Role{}, errs.Errorf(
			errs.InvalidArgument,
			errs.CodeValidation,
			"could not parse uuid: %s",
			err,
		)
	}

	rol, err := uc.roleService.QueryByID(ctx, roleUUID)
	if err != nil {
		return Role{}, errs.Errorf(errs.Internal, errs.CodeInternal, "role query by id: %s", err)
	}

	return toAppRole(rol), nil
}

// Query returns a list of roles with paging.
func (uc *UseCase) Query(ctx context.Context, qp AppQueryParams) (page.Result[Role], error) {
	p, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return page.Result[Role]{}, errs.NewFieldErrors("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return page.Result[Role]{}, err
	}

	orderBy, err := order.Parse(getOrderByFields(), qp.OrderBy, getDefaultOrderBy())
	if err != nil {
		return page.Result[Role]{}, errs.NewFieldErrors("order", err)
	}

	roles, err := uc.roleService.Query(ctx, filter, orderBy, p)
	if err != nil {
		return page.Result[Role]{}, errs.Errorf(
			errs.Internal,
			errs.CodeInternal,
			"query: %s",
			err,
		)
	}

	total, err := uc.roleService.Count(ctx, filter)
	if err != nil {
		return page.Result[Role]{}, errs.Errorf(
			errs.Internal,
			errs.CodeInternal,
			"count: %s",
			err,
		)
	}

	return page.NewResult(toAppRoles(roles), total, p), nil
}

// newRoleAudit builds an audit.NewAudit for a role operation.
func (uc *UseCase) newRoleAudit(ctx context.Context, rl role.Role, action string) audit.NewAudit {
	actorID, _ := uuid.Parse(ctxPck.GetActorID(ctx))

	return audit.NewAudit{
		ObjID:     rl.ID(),
		ObjEntity: entity.New(entity.RoleEntity),
		ObjName:   rl.Name(),
		ActorID:   actorID,
		Action:    action,
		Data:      rl,
		Message:   "role " + action,
	}
}
