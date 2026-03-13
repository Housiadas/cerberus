// Package user_roles_usecase maintains the usecase layer for user-role assignments.
package user_roles_usecase

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/audit"
	"github.com/Housiadas/cerberus/internal/core/domain/entity"
	"github.com/Housiadas/cerberus/internal/core/domain/name"
	"github.com/Housiadas/cerberus/internal/core/service/audit_service"
	"github.com/Housiadas/cerberus/internal/core/service/user_roles_service"
	ctxPck "github.com/Housiadas/cerberus/internal/utils/context"
	"github.com/Housiadas/cerberus/internal/utils/errs"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
)

// UseCase manages user-role assignment operations.
type UseCase struct {
	log              logger.Logger
	userRolesService *user_roles_service.Service
	auditSvc         *audit_service.Service
	tx               pgsql.Beginner
}

// NewUseCase constructs a UseCase.
func NewUseCase(
	log logger.Logger,
	userRolesService *user_roles_service.Service,
	auditSvc *audit_service.Service,
	tx pgsql.Beginner,
) *UseCase {
	return &UseCase{
		log:              log,
		userRolesService: userRolesService,
		auditSvc:         auditSvc,
		tx:               tx,
	}
}

// AddUserRole assigns a role to a user.
func (uc *UseCase) AddUserRole(ctx context.Context, userID, roleID string) error {
	return uc.modifyUserRole(ctx, userID, roleID, audit.ActionAssign)
}

// RemoveUserRole removes a role from a user.
func (uc *UseCase) RemoveUserRole(ctx context.Context, userID, roleID string) error {
	return uc.modifyUserRole(ctx, userID, roleID, audit.ActionRemove)
}

func (uc *UseCase) modifyUserRole(ctx context.Context, userID, roleID, action string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return errs.Errorf(
			errs.InvalidArgument,
			errs.CodeValidation,
			"could not parse user uuid: %s",
			err,
		)
	}

	roleUUID, err := uuid.Parse(roleID)
	if err != nil {
		return errs.Errorf(
			errs.InvalidArgument,
			errs.CodeValidation,
			"could not parse role uuid: %s",
			err,
		)
	}

	txErr := pgsql.RunInTx(ctx, uc.log, uc.tx, func(tran pgsql.CommitRollbacker) error {
		svcTx, initErr := uc.userRolesService.NewWithTx(tran)
		if initErr != nil {
			return errs.Errorf(errs.Internal, errs.CodeInternal, "user_roles tx: %s", initErr)
		}

		auditSvcTx, initErr := uc.auditSvc.NewWithTx(tran)
		if initErr != nil {
			return errs.Errorf(errs.Internal, errs.CodeInternal, "audit tx: %s", initErr)
		}

		var opErr error
		if action == audit.ActionAssign {
			opErr = svcTx.Add(ctx, userUUID, roleUUID)
		} else {
			opErr = svcTx.Remove(ctx, userUUID, roleUUID)
		}

		if opErr != nil {
			return errs.Errorf(
				errs.Internal,
				errs.CodeInternal,
				"user_role %s: %s",
				action,
				opErr,
			)
		}

		actorID, _ := uuid.Parse(ctxPck.GetActorID(ctx))

		_, auditErr := auditSvcTx.Create(ctx, audit.NewAudit{
			ObjID:     userUUID,
			ObjEntity: entity.New(entity.RoleEntity),
			ObjName:   name.MustParse(roleUUID.String()),
			ActorID:   actorID,
			Action:    action,
			Data:      map[string]string{"user_id": userID, "role_id": roleID},
			Message:   "user role " + action,
		})
		if auditErr != nil {
			return errs.Errorf(
				errs.Internal,
				errs.CodeInternal,
				"audit user_role %s: %s",
				action,
				auditErr,
			)
		}

		return nil
	})
	if txErr != nil {
		return fmt.Errorf("user role %s: %w", action, txErr)
	}

	return nil
}
