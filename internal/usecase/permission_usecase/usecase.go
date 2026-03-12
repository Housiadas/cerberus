// Package permission_usecase maintains the usecase layer api the model
package permission_usecase

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/audit"
	"github.com/Housiadas/cerberus/internal/core/domain/entity"
	"github.com/Housiadas/cerberus/internal/core/domain/permission"
	"github.com/Housiadas/cerberus/internal/core/service/audit_service"
	"github.com/Housiadas/cerberus/internal/core/service/permission_service"
	ctxPck "github.com/Housiadas/cerberus/internal/utils/context"
	"github.com/Housiadas/cerberus/internal/utils/errs"
	"github.com/Housiadas/cerberus/internal/utils/page"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
)

type UseCase struct {
	log               logger.Logger
	permissionService *permission_service.Service
	auditSvc          *audit_service.Service
	tx                pgsql.Beginner
}

func NewUseCase(
	log logger.Logger,
	permissionService *permission_service.Service,
	auditSvc *audit_service.Service,
	tx pgsql.Beginner,
) *UseCase {
	return &UseCase{
		log:               log,
		permissionService: permissionService,
		auditSvc:          auditSvc,
		tx:                tx,
	}
}

// Create adds a new permission to the system.
func (uc *UseCase) Create(ctx context.Context, nperm NewPermission) (Permission, error) {
	np, err := toBusNewPermission(nperm)
	if err != nil {
		return Permission{}, errs.New(errs.InvalidArgument, errs.CodeValidation, err)
	}

	var perm permission.Permission

	txErr := pgsql.RunInTx(ctx, uc.log, uc.tx, func(tran pgsql.CommitRollbacker) error {
		permServiceTx, initErr := uc.permissionService.NewWithTx(tran)
		if initErr != nil {
			return errs.Errorf(errs.Internal, errs.CodeInternal, "permission tx: %s", initErr)
		}

		auditSvcTx, initErr := uc.auditSvc.NewWithTx(tran)
		if initErr != nil {
			return errs.Errorf(errs.Internal, errs.CodeInternal, "audit tx: %s", initErr)
		}

		perm, err = permServiceTx.Create(ctx, np)
		if err != nil {
			return errs.Errorf(
				errs.Internal,
				errs.CodeInternal,
				"create: perm[%+v]: %s",
				perm,
				err,
			)
		}

		_, auditErr := auditSvcTx.Create(ctx, uc.newPermissionAudit(ctx, perm, audit.ActionCreate))
		if auditErr != nil {
			return errs.Errorf(errs.Internal, errs.CodeInternal, "audit create: %s", auditErr)
		}

		return nil
	})
	if txErr != nil {
		return Permission{}, fmt.Errorf("create permission: %w", txErr)
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
		return Permission{}, errs.New(errs.InvalidArgument, errs.CodeValidation, err)
	}

	permissionUUID, err := uuid.Parse(permissionID)
	if err != nil {
		return Permission{}, errs.Errorf(
			errs.InvalidArgument,
			errs.CodeValidation,
			"could not parse uuid: %s",
			err,
		)
	}

	currentPerm, err := uc.permissionService.QueryByID(ctx, permissionUUID)
	if err != nil {
		return Permission{}, errs.Errorf(
			errs.Internal,
			errs.CodeInternal,
			"permission query by id: %s",
			err,
		)
	}

	var updPerm permission.Permission

	txErr := pgsql.RunInTx(ctx, uc.log, uc.tx, func(tran pgsql.CommitRollbacker) error {
		permServiceTx, initErr := uc.permissionService.NewWithTx(tran)
		if initErr != nil {
			return errs.Errorf(errs.Internal, errs.CodeInternal, "permission tx: %s", initErr)
		}

		auditSvcTx, initErr := uc.auditSvc.NewWithTx(tran)
		if initErr != nil {
			return errs.Errorf(errs.Internal, errs.CodeInternal, "audit tx: %s", initErr)
		}

		updPerm, err = permServiceTx.Update(ctx, currentPerm, up)
		if err != nil {
			return errs.Errorf(
				errs.Internal,
				errs.CodeInternal,
				"update: permissionID[%s] up[%+v]: %s",
				permissionID,
				up,
				err,
			)
		}

		_, auditErr := auditSvcTx.Create(
			ctx,
			uc.newPermissionAudit(ctx, updPerm, audit.ActionUpdate),
		)
		if auditErr != nil {
			return errs.Errorf(errs.Internal, errs.CodeInternal, "audit update: %s", auditErr)
		}

		return nil
	})
	if txErr != nil {
		return Permission{}, fmt.Errorf("update permission: %w", txErr)
	}

	return toAppPermission(updPerm), nil
}

// Delete removes a permission from the system.
func (uc *UseCase) Delete(ctx context.Context, permissionID string) error {
	permissionUUID, err := uuid.Parse(permissionID)
	if err != nil {
		return errs.Errorf(
			errs.InvalidArgument,
			errs.CodeValidation,
			"could not parse uuid: %s",
			err,
		)
	}

	currentPerm, err := uc.permissionService.QueryByID(ctx, permissionUUID)
	if err != nil {
		return errs.Errorf(errs.Internal, errs.CodeInternal, "permission query by id: %s", err)
	}

	txErr := pgsql.RunInTx(ctx, uc.log, uc.tx, func(tran pgsql.CommitRollbacker) error {
		permServiceTx, initErr := uc.permissionService.NewWithTx(tran)
		if initErr != nil {
			return errs.Errorf(errs.Internal, errs.CodeInternal, "permission tx: %s", initErr)
		}

		auditSvcTx, initErr := uc.auditSvc.NewWithTx(tran)
		if initErr != nil {
			return errs.Errorf(errs.Internal, errs.CodeInternal, "audit tx: %s", initErr)
		}

		deleteErr := permServiceTx.Delete(ctx, currentPerm)
		if deleteErr != nil {
			return errs.Errorf(
				errs.Internal,
				errs.CodeInternal,
				"delete: permissionID[%s]: %s",
				permissionUUID,
				deleteErr,
			)
		}

		_, auditErr := auditSvcTx.Create(
			ctx,
			uc.newPermissionAudit(ctx, currentPerm, audit.ActionDelete),
		)
		if auditErr != nil {
			return errs.Errorf(errs.Internal, errs.CodeInternal, "audit delete: %s", auditErr)
		}

		return nil
	})
	if txErr != nil {
		return fmt.Errorf("delete permission: %w", txErr)
	}

	return nil
}

// QueryByID returns a permission by its ID.
func (uc *UseCase) QueryByID(ctx context.Context, permissionID string) (Permission, error) {
	permissionUUID, err := uuid.Parse(permissionID)
	if err != nil {
		return Permission{}, errs.Errorf(
			errs.InvalidArgument,
			errs.CodeValidation,
			"could not parse uuid: %s",
			err,
		)
	}

	perm, err := uc.permissionService.QueryByID(ctx, permissionUUID)
	if err != nil {
		return Permission{}, errs.Errorf(
			errs.Internal,
			errs.CodeInternal,
			"permission query by id: %s",
			err,
		)
	}

	return toAppPermission(perm), nil
}

// Query returns a list of permissions with paging.
func (uc *UseCase) Query(ctx context.Context, qp AppQueryParams) (page.Result[Permission], error) {
	p, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return page.Result[Permission]{}, errs.NewFieldErrors("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return page.Result[Permission]{}, err
	}

	orderBy, err := order.Parse(getOrderByFields(), qp.OrderBy, getDefaultOrderBy())
	if err != nil {
		return page.Result[Permission]{}, errs.NewFieldErrors("order", err)
	}

	perms, err := uc.permissionService.Query(ctx, filter, orderBy, p)
	if err != nil {
		return page.Result[Permission]{}, errs.Errorf(
			errs.Internal,
			errs.CodeInternal,
			"query: %s",
			err,
		)
	}

	total, err := uc.permissionService.Count(ctx, filter)
	if err != nil {
		return page.Result[Permission]{}, errs.Errorf(
			errs.Internal,
			errs.CodeInternal,
			"count: %s",
			err,
		)
	}

	return page.NewResult(toAppPermissions(perms), total, p), nil
}

// newPermissionAudit builds an audit.NewAudit for a permission operation.
func (uc *UseCase) newPermissionAudit(
	ctx context.Context,
	p permission.Permission,
	action string,
) audit.NewAudit {
	actorID, _ := uuid.Parse(ctxPck.GetActorID(ctx))

	return audit.NewAudit{
		ObjID:     p.ID(),
		ObjEntity: entity.New(entity.PermissionEntity),
		ObjName:   p.Name(),
		ActorID:   actorID,
		Action:    action,
		Data:      p,
		Message:   "permission " + action,
	}
}
