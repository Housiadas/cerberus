// Package eventbus provides a generic domain event dispatcher
// that persists events to the outbox and audit log within a transaction.
package eventbus

import (
	"context"

	ctxPck "github.com/Housiadas/cerberus/internal/context"
	"github.com/Housiadas/cerberus/internal/core/domain/audit"
	"github.com/Housiadas/cerberus/internal/core/domain/outbox"
	"github.com/Housiadas/cerberus/internal/core/service/audit_service"
	"github.com/Housiadas/cerberus/internal/core/service/outbox_service"
	errs2 "github.com/Housiadas/cerberus/internal/errs"
	"github.com/Housiadas/cerberus/internal/types/event"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
)

// EventDispatcher dispatches domain events to the outbox and audit log.
type EventDispatcher struct {
	outboxSvc *outbox_service.Service
	auditSvc  *audit_service.Service
}

// New constructs an EventDispatcher.
func New(
	outboxSvc *outbox_service.Service,
	auditSvc *audit_service.Service,
) *EventDispatcher {
	return &EventDispatcher{
		outboxSvc: outboxSvc,
		auditSvc:  auditSvc,
	}
}

// Dispatch persists the given domain events within the provided transaction.
func (d *EventDispatcher) Dispatch(
	ctx context.Context,
	tran pgsql.CommitRollbacker,
	ev event.DomainEvent,
) error {
	outboxSvcTx, err := d.outboxSvc.NewWithTx(tran)
	if err != nil {
		return errs2.Errorf(errs2.Internal, errs2.CodeInternal, "outbox tx: %s", err)
	}

	auditSvcTx, err := d.auditSvc.NewWithTx(tran)
	if err != nil {
		return errs2.Errorf(errs2.Internal, errs2.CodeInternal, "audit tx: %s", err)
	}

	actorID, _ := uuid.Parse(ctxPck.GetActorID(ctx))

	// if a topic is present, write to outbox table to produce to kafka
	if ev.Topic != "" {
		outboxErr := outboxSvcTx.Create(ctx, outbox.NewOutbox{
			EventType:   ev.EventType,
			AggregateID: ev.AggregateID,
			Topic:       ev.Topic,
			Payload:     ev.Payload,
		})
		if outboxErr != nil {
			return errs2.Errorf(
				errs2.Internal,
				errs2.CodeInternal,
				"outbox create: %s",
				outboxErr,
			)
		}
	}

	// Add audit log
	_, auditErr := auditSvcTx.Create(ctx, audit.NewAudit{
		ObjID:     ev.AggregateID,
		ObjEntity: ev.ObjEntity,
		ObjName:   ev.ObjName,
		ActorID:   actorID,
		Action:    ev.Action,
		Data:      ev.Payload,
		Message:   ev.Message,
	})
	if auditErr != nil {
		return errs2.Errorf(
			errs2.Internal,
			errs2.CodeInternal,
			"audit create: %s",
			auditErr,
		)
	}

	return nil
}
