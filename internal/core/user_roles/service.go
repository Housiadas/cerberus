// Package user_roles manages user role assignments.
package user_roles

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

// Service manages user role assignments.
type Service struct {
	log        logger.Logger
	storer     Storer
	tx         transactor
	dispatcher dispatcher
}

// NewService constructs a business API for use.
func NewService(
	log logger.Logger,
	storer Storer,
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

// Add assigns a role to a user within a transaction and dispatches a domain event.
func (s *Service) Add(ctx context.Context, userID uuid.UUID, roleID uuid.UUID) error {
	return s.modify(ctx, userID, roleID, audit.ActionAssign)
}

// Remove removes a role from a user within a transaction and dispatches a domain event.
func (s *Service) Remove(ctx context.Context, userID uuid.UUID, roleID uuid.UUID) error {
	return s.modify(ctx, userID, roleID, audit.ActionRemove)
}

func (s *Service) modify(ctx context.Context, userID, roleID uuid.UUID, action string) error {
	txErr := s.tx.RunInTx(ctx, func(txCtx context.Context) error {
		var opErr error
		if action == audit.ActionAssign {
			opErr = s.storer.Add(txCtx, userID, roleID)
		} else {
			opErr = s.storer.Remove(txCtx, userID, roleID)
		}

		if opErr != nil {
			return fmt.Errorf("user_role %s: %w", action, opErr)
		}

		return s.dispatcher.Dispatch(txCtx, event.DomainEvent{
			AggregateID: userID,
			Payload:     map[string]string{"user_id": userID.String(), "role_id": roleID.String()},
			ObjEntity:   entity.New(entity.RoleEntity),
			ObjName:     name.MustParse(roleID.String()),
			Action:      action,
			Message:     "user role " + action,
		})
	})
	if txErr != nil {
		return fmt.Errorf("user role %s: %w", action, txErr)
	}

	return nil
}
