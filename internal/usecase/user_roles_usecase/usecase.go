// Package user_roles_usecase maintains the usecase layer for user-role assignments.
package user_roles_usecase

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/audit"
	"github.com/Housiadas/cerberus/internal/core/service/user_roles_service"
	errs2 "github.com/Housiadas/cerberus/internal/errs"
	"github.com/Housiadas/cerberus/internal/eventbus"
	"github.com/Housiadas/cerberus/internal/types/entity"
	"github.com/Housiadas/cerberus/internal/types/event"
	"github.com/Housiadas/cerberus/internal/types/name"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
)

// UseCase manages user-role assignment operations.
type UseCase struct {
	log              logger.Logger
	userRolesService *user_roles_service.Service
	dispatcher       *eventbus.EventDispatcher
	tx               pgsql.Beginner
}

// NewUseCase constructs a UseCase.
func NewUseCase(
	log logger.Logger,
	userRolesService *user_roles_service.Service,
	dispatcher *eventbus.EventDispatcher,
	tx pgsql.Beginner,
) *UseCase {
	return &UseCase{
		log:              log,
		userRolesService: userRolesService,
		dispatcher:       dispatcher,
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
		return errs2.Errorf(
			errs2.InvalidArgument,
			errs2.CodeValidation,
			"could not parse user uuid: %s",
			err,
		)
	}

	roleUUID, err := uuid.Parse(roleID)
	if err != nil {
		return errs2.Errorf(
			errs2.InvalidArgument,
			errs2.CodeValidation,
			"could not parse role uuid: %s",
			err,
		)
	}

	txErr := pgsql.RunInTx(ctx, uc.log, uc.tx, func(tran pgsql.CommitRollbacker) error {
		svcTx, initErr := uc.userRolesService.NewWithTx(tran)
		if initErr != nil {
			return errs2.Errorf(errs2.Internal, errs2.CodeInternal, "user_roles tx: %s", initErr)
		}

		var opErr error
		if action == audit.ActionAssign {
			opErr = svcTx.Add(ctx, userUUID, roleUUID)
		} else {
			opErr = svcTx.Remove(ctx, userUUID, roleUUID)
		}

		if opErr != nil {
			return errs2.Errorf(
				errs2.Internal,
				errs2.CodeInternal,
				"user_role %s: %s",
				action,
				opErr,
			)
		}

		return uc.dispatcher.Dispatch(ctx, tran, event.DomainEvent{
			AggregateID: userUUID,
			Payload:     map[string]string{"user_id": userID, "role_id": roleID},
			ObjEntity:   entity.New(entity.RoleEntity),
			ObjName:     name.MustParse(roleUUID.String()),
			Action:      action,
			Message:     "user role " + action,
		})
	})
	if txErr != nil {
		return fmt.Errorf("user role %s: %w", action, txErr)
	}

	return nil
}
