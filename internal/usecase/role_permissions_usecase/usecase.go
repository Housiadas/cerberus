// Package role_permissions_usecase maintains the usecase layer for role-permission assignments.
package role_permissions_usecase

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/audit"
	"github.com/Housiadas/cerberus/internal/core/service/role_permissions_service"
	errs2 "github.com/Housiadas/cerberus/internal/errs"
	"github.com/Housiadas/cerberus/internal/eventbus"
	"github.com/Housiadas/cerberus/internal/types/entity"
	"github.com/Housiadas/cerberus/internal/types/event"
	"github.com/Housiadas/cerberus/internal/types/name"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
)

// UseCase manages role-permission assignment operations.
type UseCase struct {
	log              logger.Logger
	tx               pgsql.Beginner
	dispatcher       *eventbus.EventDispatcher
	rolePermsService *role_permissions_service.Service
}

// NewUseCase constructs a UseCase.
func NewUseCase(
	log logger.Logger,
	rolePermsService *role_permissions_service.Service,
	dispatcher *eventbus.EventDispatcher,
	tx pgsql.Beginner,
) *UseCase {
	return &UseCase{
		log:              log,
		tx:               tx,
		dispatcher:       dispatcher,
		rolePermsService: rolePermsService,
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
		return errs2.Errorf(
			errs2.InvalidArgument,
			errs2.CodeValidation,
			"could not parse role uuid: %s",
			err,
		)
	}

	permUUID, err := uuid.Parse(permissionID)
	if err != nil {
		return errs2.Errorf(
			errs2.InvalidArgument,
			errs2.CodeValidation,
			"could not parse permission uuid: %s",
			err,
		)
	}

	txErr := pgsql.RunInTx(ctx, uc.log, uc.tx, func(tran pgsql.CommitRollbacker) error {
		rolePermsTx, initErr := uc.rolePermsService.NewWithTx(tran)
		if initErr != nil {
			return errs2.Errorf(
				errs2.Internal,
				errs2.CodeInternal,
				"role_permissions tx: %s",
				initErr,
			)
		}

		var opErr error
		if action == audit.ActionAssign {
			opErr = rolePermsTx.Add(ctx, roleUUID, permUUID)
		} else {
			opErr = rolePermsTx.Remove(ctx, roleUUID, permUUID)
		}

		if opErr != nil {
			return errs2.Errorf(
				errs2.Internal,
				errs2.CodeInternal,
				"role_permission %s: %s",
				action,
				opErr,
			)
		}

		return uc.dispatcher.Dispatch(ctx, tran, event.DomainEvent{
			AggregateID: roleUUID,
			Payload:     map[string]string{"role_id": roleID, "permission_id": permissionID},
			ObjEntity:   entity.New(entity.PermissionEntity),
			ObjName:     name.MustParse(permUUID.String()),
			Action:      action,
			Message:     "role permission " + action,
		})
	})
	if txErr != nil {
		return fmt.Errorf("role permission %s: %w", action, txErr)
	}

	return nil
}
