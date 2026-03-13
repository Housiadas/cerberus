package event

import (
	"github.com/Housiadas/cerberus/internal/core/domain/entity"
	"github.com/Housiadas/cerberus/internal/core/domain/name"
	"github.com/google/uuid"
)

// DomainEvent represents a domain event that should be dispatched
// to the outbox (if Topic is non-empty) and always to the audit log.
type DomainEvent struct {
	EventType   string
	AggregateID uuid.UUID
	Topic       string
	Payload     any
	ObjEntity   entity.Entity
	ObjName     name.Name
	Action      string
	Message     string
}
