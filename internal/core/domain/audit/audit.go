package audit

import (
	"encoding/json"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/entity"
	"github.com/Housiadas/cerberus/internal/core/domain/name"
	"github.com/google/uuid"
)

// Audit represents information about an individual audit record.
type Audit struct {
	id        uuid.UUID
	objID     uuid.UUID
	objEntity entity.Entity
	objName   name.Name
	actorID   uuid.UUID
	action    string
	data      json.RawMessage
	message   string
	timestamp time.Time
}

// New constructs an Audit from its individual fields.
func New(
	id uuid.UUID,
	objID uuid.UUID,
	objEntity entity.Entity,
	objName name.Name,
	actorID uuid.UUID,
	action string,
	data json.RawMessage,
	message string,
	timestamp time.Time,
) Audit {
	return Audit{
		id:        id,
		objID:     objID,
		objEntity: objEntity,
		objName:   objName,
		actorID:   actorID,
		action:    action,
		data:      data,
		message:   message,
		timestamp: timestamp,
	}
}

// ID returns the audit ID.
func (a Audit) ID() uuid.UUID { return a.id }

// ObjID returns the audited object ID.
func (a Audit) ObjID() uuid.UUID { return a.objID }

// ObjEntity returns the audited object entity type.
func (a Audit) ObjEntity() entity.Entity { return a.objEntity }

// ObjName returns the audited object name.
func (a Audit) ObjName() name.Name { return a.objName }

// ActorID returns the actor ID.
func (a Audit) ActorID() uuid.UUID { return a.actorID }

// Action returns the action performed.
func (a Audit) Action() string { return a.action }

// Data returns the audit data payload.
func (a Audit) Data() json.RawMessage { return a.data }

// Message returns the audit message.
func (a Audit) Message() string { return a.message }

// Timestamp returns the audit timestamp.
func (a Audit) Timestamp() time.Time { return a.timestamp }

// NewAudit represents the information needed to create a new audit record.
type NewAudit struct {
	ObjID     uuid.UUID
	ObjEntity entity.Entity
	ObjName   name.Name
	ActorID   uuid.UUID
	Action    string
	Data      any
	Message   string
}

// QueryFilter holds the available fields a query can be filtered on.
// We are using pointer semantics because the With API mutates the value.
type QueryFilter struct {
	ObjID     *uuid.UUID
	ObjEntity *entity.Entity
	ObjName   *name.Name
	ActorID   *uuid.UUID
	Action    *string
	Since     *time.Time
	Until     *time.Time
}
