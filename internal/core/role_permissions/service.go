// Package role_permissions manages role permission assignments.
package role_permissions

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/audit"
	"github.com/Housiadas/cerberus/internal/types/entity"
	"github.com/Housiadas/cerberus/internal/types/event"
	"github.com/Housiadas/cerberus/internal/types/name"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/google/uuid"
)

// Service manages role permission assignments.
type Service struct {
	log        logger.Logger
	storer     storer
	tx         transactor
	dispatcher dispatcher
}

// NewService constructs a business API for use.
func NewService(
	log logger.Logger,
	storer storer,
	tx transactor,
	dispatcher dispatcher,
) *Service {
	return &Service{
		log:        log,
		storer:     storer,
		tx:         tx,
		dispatcher: dispatcher,
	}
}

// Add assigns a permission to a role within a transaction and dispatches a domain event.
func (s *Service) Add(ctx context.Context, roleID uuid.UUID, permissionID uuid.UUID) error {
	return s.modify(ctx, roleID, permissionID, audit.ActionAssign)
}

// Remove removes a permission from a role within a transaction and dispatches a domain event.
func (s *Service) Remove(ctx context.Context, roleID uuid.UUID, permissionID uuid.UUID) error {
	return s.modify(ctx, roleID, permissionID, audit.ActionRemove)
}

func (s *Service) modify(ctx context.Context, roleID, permissionID uuid.UUID, action string) error {
	txErr := s.tx.RunInTx(ctx, func(txCtx context.Context) error {
		var opErr error
		if action == audit.ActionAssign {
			params := toCreateRolePermissionParams(roleID, permissionID)
			_, opErr = s.storer.CreateRolePermission(txCtx, params)
		} else {
			params := toDeleteRolePermissionParams(roleID, permissionID)
			opErr = s.storer.DeleteRolePermission(txCtx, params)
		}

		if opErr != nil {
			return fmt.Errorf("role_permission %s: %w", action, opErr)
		}

		return s.dispatcher.Dispatch(txCtx, event.DomainEvent{
			AggregateID: roleID,
			Payload: map[string]string{
				"role_id":       roleID.String(),
				"permission_id": permissionID.String(),
			},
			ObjEntity: entity.New(entity.PermissionEntity),
			ObjName:   name.MustParse(permissionID.String()),
			Action:    action,
			Message:   "role permission " + action,
		})
	})
	if txErr != nil {
		return fmt.Errorf("role permission %s: %w", action, txErr)
	}

	return nil
}
