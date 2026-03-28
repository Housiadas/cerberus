// Package role_permissions manages role permission assignments.
package role_permissions

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/audit"
	"github.com/Housiadas/cerberus/internal/types/entity"
	"github.com/Housiadas/cerberus/internal/types/event"
	"github.com/Housiadas/cerberus/internal/types/name"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type logger interface {
	Info(ctx context.Context, msg string, args ...any)
	Infoc(ctx context.Context, caller int, msg string, args ...any)
	Warn(ctx context.Context, msg string, args ...any)
	Warnc(ctx context.Context, caller int, msg string, args ...any)
	Error(ctx context.Context, msg string, args ...any)
	Errorc(ctx context.Context, caller int, msg string, args ...any)
}

// dispatcher defines the interface for domain event dispatching.
type dispatcher interface {
	Dispatch(ctx context.Context, ev event.DomainEvent) error
}

// Service manages role permission assignments.
type Service struct {
	log        logger
	storer     Storer
	db         *sqlx.DB
	dispatcher dispatcher
}

// NewService constructs a business API for use.
func NewService(
	log logger,
	storer Storer,
	db *sqlx.DB,
	dispatcher dispatcher,
) *Service {
	return &Service{
		log:        log,
		storer:     storer,
		db:         db,
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
	txErr := pgsql.RunInTx(ctx, s.log, s.db, func(txCtx context.Context) error {
		var opErr error
		if action == audit.ActionAssign {
			opErr = s.storer.Add(txCtx, roleID, permissionID)
		} else {
			opErr = s.storer.Remove(txCtx, roleID, permissionID)
		}

		if opErr != nil {
			return fmt.Errorf("role_permission %s: %w", action, opErr)
		}

		return s.dispatcher.Dispatch(txCtx, event.DomainEvent{
			AggregateID: roleID,
			Payload:     map[string]string{"role_id": roleID.String(), "permission_id": permissionID.String()},
			ObjEntity:   entity.New(entity.PermissionEntity),
			ObjName:     name.MustParse(permissionID.String()),
			Action:      action,
			Message:     "role permission " + action,
		})
	})
	if txErr != nil {
		return fmt.Errorf("role permission %s: %w", action, txErr)
	}

	return nil
}
