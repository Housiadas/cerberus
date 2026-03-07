// Package role_permissions_usecase maintains the usecase layer for role-permission assignments.
package role_permissions_usecase

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/audit"
	"github.com/Housiadas/cerberus/internal/core/domain/entity"
	"github.com/Housiadas/cerberus/internal/core/domain/name"
	"github.com/Housiadas/cerberus/internal/core/service/audit_service"
	"github.com/Housiadas/cerberus/internal/core/service/role_permissions_service"
	ctxPck "github.com/Housiadas/cerberus/internal/utils/context"
	"github.com/Housiadas/cerberus/internal/utils/errs"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
)

// UseCase manages role-permission assignment operations.
type UseCase struct {
	log              logger.Logger
	rolePermsService *role_permissions_service.Service
	auditSvc         *audit_service.Service
	tx               pgsql.Beginner
}

// NewUseCase constructs a UseCase.
func NewUseCase(
	log logger.Logger,
	rolePermsService *role_permissions_service.Service,
	auditSvc *audit_service.Service,
	tx pgsql.Beginner,
) *UseCase {
	return &UseCase{
		log:              log,
		rolePermsService: rolePermsService,
		auditSvc:         auditSvc,
		tx:               tx,
	}
}

// AddRolePermission assigns a permission to a role.
func (uc *UseCase) AddRolePermission(ctx context.Context, roleID, permissionID string) error {
	return uc.modifyRolePermission(ctx, roleID, permissionID, audit.ActionAssign)
}

// RemoveRolePermission removes a permission from a role.
func (uc *UseCase) RemoveRolePermission(ctx context.Context, roleID, permissionID string) error {
	return uc.modifyRolePermission(ctx, roleID, permissionID, audit.ActionRemove)
}

func (uc *UseCase) modifyRolePermission(
	ctx context.Context,
	roleID, permissionID, action string,
) error {
	roleUUID, err := uuid.Parse(roleID)
	if err != nil {
		return errs.Errorf(errs.InvalidArgument, "could not parse role uuid: %s", err)
	}

	permUUID, err := uuid.Parse(permissionID)
	if err != nil {
		return errs.Errorf(errs.InvalidArgument, "could not parse permission uuid: %s", err)
	}

	txErr := pgsql.RunInTx(ctx, uc.log, uc.tx, func(tran pgsql.CommitRollbacker) error {
		rolePermsTx, initErr := uc.rolePermsService.NewWithTx(tran)
		if initErr != nil {
			return errs.Errorf(errs.Internal, "role_permissions tx: %s", initErr)
		}

		auditSvcTx, initErr := uc.auditSvc.NewWithTx(tran)
		if initErr != nil {
			return errs.Errorf(errs.Internal, "audit tx: %s", initErr)
		}

		var opErr error
		if action == audit.ActionAssign {
			opErr = rolePermsTx.Add(ctx, roleUUID, permUUID)
		} else {
			opErr = rolePermsTx.Remove(ctx, roleUUID, permUUID)
		}

		if opErr != nil {
			return errs.Errorf(errs.Internal, "role_permission %s: %s", action, opErr)
		}

		actorID, _ := uuid.Parse(ctxPck.GetActorID(ctx))

		_, auditErr := auditSvcTx.Create(ctx, audit.NewAudit{
			ObjID:     roleUUID,
			ObjEntity: entity.New(entity.PermissionEntity),
			ObjName:   name.MustParse(permUUID.String()),
			ActorID:   actorID,
			Action:    action,
			Data:      map[string]string{"role_id": roleID, "permission_id": permissionID},
			Message:   "role permission " + action,
		})
		if auditErr != nil {
			return errs.Errorf(errs.Internal, "audit role_permission %s: %s", action, auditErr)
		}

		return nil
	})
	if txErr != nil {
		return fmt.Errorf("role permission %s: %w", action, txErr)
	}

	return nil
}
