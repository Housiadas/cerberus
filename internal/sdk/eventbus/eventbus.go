// Package eventbus provides a generic domain event dispatcher
// that persists events to the outbox and audit log within a transaction.
package eventbus

import (
	"context"

	"github.com/Housiadas/cerberus/internal/core/audit"
	"github.com/Housiadas/cerberus/internal/core/outbox"
	ctxPck "github.com/Housiadas/cerberus/internal/sdk/context"
	errs "github.com/Housiadas/cerberus/internal/sdk/errs"
	"github.com/Housiadas/cerberus/internal/types/event"
	"github.com/google/uuid"
)

type outboxer interface {
	Create(ctx context.Context, outbox outbox.NewOutbox) error
}

type auditer interface {
	Create(ctx context.Context, audit audit.NewAudit) (audit.Audit, error)
}

// EventDispatcher dispatches domain events to the outbox and audit log.
type EventDispatcher struct {
	outboxSvc outboxer
	auditSvc  auditer
}

// New constructs an EventDispatcher.
func New(
	outboxSvc outboxer,
	auditSvc auditer,
) *EventDispatcher {
	return &EventDispatcher{
		outboxSvc: outboxSvc,
		auditSvc:  auditSvc,
	}
}

// Dispatch persists the given domain events within the provided transaction context.
func (d *EventDispatcher) Dispatch(
	ctx context.Context,
	ev event.DomainEvent,
) error {
	actorID, _ := uuid.Parse(ctxPck.GetActorID(ctx))

	// if a topic is present, write to outbox table to produce to kafka
	if ev.Topic != "" {
		outboxErr := d.outboxSvc.Create(ctx, outbox.NewOutbox{
			EventType:   ev.EventType,
			AggregateID: ev.AggregateID,
			Topic:       ev.Topic,
			Payload:     ev.Payload,
		})
		if outboxErr != nil {
			return errs.Errorf(
				errs.Internal,
				errs.CodeInternal,
				"outbox create: %s",
				outboxErr,
			)
		}
	}

	// Add audit log
	_, auditErr := d.auditSvc.Create(ctx, audit.NewAudit{
		ObjID:     ev.AggregateID,
		ObjEntity: ev.ObjEntity,
		ObjName:   ev.ObjName,
		ActorID:   actorID,
		Action:    ev.Action,
		Data:      ev.Payload,
		Message:   ev.Message,
	})
	if auditErr != nil {
		return errs.Errorf(
			errs.Internal,
			errs.CodeInternal,
			"audit create: %s",
			auditErr,
		)
	}

	return nil
}
