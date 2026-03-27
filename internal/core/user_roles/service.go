// Package user_roles manages user role assignments.
package user_roles

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

// Service manages user role assignments.
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
		return nil, fmt.Errorf("user_roles transaction issue: %w", err)
	}

	return &Service{
		log:        s.log,
		storer:     storer,
		tx:         s.tx,
		dispatcher: s.dispatcher,
	}, nil
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
	txErr := pgsql.RunInTx(ctx, s.log, s.tx, func(tran pgsql.CommitRollbacker) error {
		storerTx, err := s.storer.NewWithTx(tran)
		if err != nil {
			return fmt.Errorf("user_roles tx: %w", err)
		}

		var opErr error
		if action == audit.ActionAssign {
			opErr = storerTx.Add(ctx, userID, roleID)
		} else {
			opErr = storerTx.Remove(ctx, userID, roleID)
		}

		if opErr != nil {
			return fmt.Errorf("user_role %s: %w", action, opErr)
		}

		return s.dispatcher.Dispatch(ctx, tran, event.DomainEvent{
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
