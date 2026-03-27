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
	Dispatch(ctx context.Context, tran pgsql.CommitRollbacker, ev event.DomainEvent) error
}

// Service manages role permission assignments.
type Service struct {
	log        logger
	storer     Storer
	tx         pgsql.Beginner
	dispatcher dispatcher
}

// NewService constructs a business API for use.
func NewService(
	log logger,
	storer Storer,
	tx pgsql.Beginner,
	dispatcher dispatcher,
) *Service {
	return &Service{
		log:        log,
		storer:     storer,
		tx:         tx,
		dispatcher: dispatcher,
	}
}

// NewWithTx constructs a new Service value replacing the sqlx DB
// value with a sqlx DB value that is currently inside a transaction.
func (s *Service) NewWithTx(tx pgsql.CommitRollbacker) (*Service, error) {
	storer, err := s.storer.NewWithTx(tx)
	if err != nil {
		return nil, fmt.Errorf("role_permissions transaction issue: %w", err)
	}

	return &Service{
		log:        s.log,
		storer:     storer,
		tx:         s.tx,
		dispatcher: s.dispatcher,
	}, nil
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
	txErr := pgsql.RunInTx(ctx, s.log, s.tx, func(tran pgsql.CommitRollbacker) error {
		stTx, err := s.storer.NewWithTx(tran)
		if err != nil {
			return fmt.Errorf("role_permissions tx: %w", err)
		}

		var opErr error
		if action == audit.ActionAssign {
			opErr = stTx.Add(ctx, roleID, permissionID)
		} else {
			opErr = stTx.Remove(ctx, roleID, permissionID)
		}

		if opErr != nil {
			return fmt.Errorf("role_permission %s: %w", action, opErr)
		}

		return s.dispatcher.Dispatch(ctx, tran, event.DomainEvent{
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
