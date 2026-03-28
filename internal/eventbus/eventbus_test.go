package eventbus_test

import (
	"context"
	"errors"
	"testing"

	ctxPck "github.com/Housiadas/cerberus/internal/context"
	"github.com/Housiadas/cerberus/internal/core/audit"
	"github.com/Housiadas/cerberus/internal/eventbus"
	"github.com/Housiadas/cerberus/internal/types/entity"
	"github.com/Housiadas/cerberus/internal/types/event"
	"github.com/Housiadas/cerberus/internal/types/name"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDispatch_WithTopicCreatesOutboxAndAudit(t *testing.T) {
	actorUUID := uuid.MustParse("aaaaaaaa-aaaa-7aaa-aaaa-aaaaaaaaaaaa")
	ctx := ctxPck.SetActorID(context.Background(), actorUUID.String())

	ev := event.DomainEvent{
		EventType:   event.UserCreated,
		AggregateID: uuid.MustParse("22222222-2222-7222-2222-222222222222"),
		Topic:       event.UserTopic,
		Payload:     map[string]string{"name": "John"},
		ObjEntity:   entity.New(entity.UserEntity),
		ObjName:     name.MustParse("John"),
		Action:      audit.ActionCreate,
		Message:     "user CREATE",
	}

	mOutboxSvc := newMockoutboxer(t)
	mOutboxSvc.EXPECT().Create(ctx, mock.AnythingOfType("outbox.NewOutbox")).Return(nil)

	mAuditSvc := newMockauditer(t)
	mAuditSvc.EXPECT().Create(ctx, mock.AnythingOfType("audit.NewAudit")).Return(audit.Audit{}, nil)

	dispatcher := eventbus.New(mOutboxSvc, mAuditSvc)
	err := dispatcher.Dispatch(ctx, ev)
	assert.NoError(t, err)
}

func TestDispatch_WithoutTopicSkipsOutbox(t *testing.T) {
	actorUUID := uuid.MustParse("aaaaaaaa-aaaa-7aaa-aaaa-aaaaaaaaaaaa")
	ctx := ctxPck.SetActorID(context.Background(), actorUUID.String())

	ev := event.DomainEvent{
		AggregateID: uuid.MustParse("33333333-3333-7333-3333-333333333333"),
		Payload:     map[string]string{"name": "Admin"},
		ObjEntity:   entity.New(entity.RoleEntity),
		ObjName:     name.MustParse("Admin"),
		Action:      audit.ActionCreate,
		Message:     "role CREATE",
	}

	mOutboxSvc := newMockoutboxer(t)
	mAuditSvc := newMockauditer(t)
	mAuditSvc.EXPECT().Create(ctx, mock.AnythingOfType("audit.NewAudit")).Return(audit.Audit{}, nil)

	dispatcher := eventbus.New(mOutboxSvc, mAuditSvc)
	err := dispatcher.Dispatch(ctx, ev)
	assert.NoError(t, err)
}

func TestDispatch_MissingActorIDUsesZeroUUID(t *testing.T) {
	ctx := context.Background()

	ev := event.DomainEvent{
		AggregateID: uuid.MustParse("44444444-4444-7444-4444-444444444444"),
		Payload:     map[string]string{"test": "data"},
		ObjEntity:   entity.New(entity.UserEntity),
		ObjName:     name.MustParse("test"),
		Action:      audit.ActionCreate,
		Message:     "test",
	}

	mOutboxSvc := newMockoutboxer(t)
	mAuditSvc := newMockauditer(t)
	mAuditSvc.EXPECT().Create(ctx, mock.AnythingOfType("audit.NewAudit")).Return(audit.Audit{}, nil)

	dispatcher := eventbus.New(mOutboxSvc, mAuditSvc)
	err := dispatcher.Dispatch(ctx, ev)
	assert.NoError(t, err)
}

func TestDispatch_OutboxCreateError(t *testing.T) {
	ctx := context.Background()

	ev := event.DomainEvent{
		EventType:   event.UserCreated,
		AggregateID: uuid.New(),
		Topic:       event.UserTopic,
		Payload:     map[string]string{"test": "data"},
		ObjEntity:   entity.New(entity.UserEntity),
		ObjName:     name.MustParse("test"),
		Action:      audit.ActionCreate,
		Message:     "test",
	}

	mOutboxSvc := newMockoutboxer(t)
	mOutboxSvc.EXPECT().Create(ctx, mock.AnythingOfType("outbox.NewOutbox")).Return(errors.New("db error"))

	mAuditSvc := newMockauditer(t)

	dispatcher := eventbus.New(mOutboxSvc, mAuditSvc)
	err := dispatcher.Dispatch(ctx, ev)
	assert.Error(t, err)
}

func TestDispatch_AuditCreateError(t *testing.T) {
	ctx := context.Background()

	ev := event.DomainEvent{
		AggregateID: uuid.New(),
		Payload:     map[string]string{"test": "data"},
		ObjEntity:   entity.New(entity.UserEntity),
		ObjName:     name.MustParse("test"),
		Action:      audit.ActionCreate,
		Message:     "test",
	}

	mOutboxSvc := newMockoutboxer(t)

	mAuditSvc := newMockauditer(t)
	mAuditSvc.EXPECT().Create(ctx, mock.AnythingOfType("audit.NewAudit")).Return(audit.Audit{}, errors.New("audit error"))

	dispatcher := eventbus.New(mOutboxSvc, mAuditSvc)
	err := dispatcher.Dispatch(ctx, ev)
	assert.Error(t, err)
}
