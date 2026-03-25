package eventbus_test

import (
	"context"
	"testing"
	"time"

	ctxPck "github.com/Housiadas/cerberus/internal/context"
	"github.com/Housiadas/cerberus/internal/core/service/outbox_service"
	"github.com/Housiadas/cerberus/internal/eventbus"
	"github.com/Housiadas/cerberus/internal/types/entity"
	"github.com/Housiadas/cerberus/internal/types/event"
	"github.com/Housiadas/cerberus/internal/types/name"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Housiadas/cerberus/internal/core/domain/audit"
	"github.com/Housiadas/cerberus/internal/core/domain/outbox"
	"github.com/Housiadas/cerberus/internal/core/service/audit_service"
	"github.com/Housiadas/cerberus/pkg/clock"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
)

type mockTx struct{}

func (m *mockTx) Commit() error   { return nil }
func (m *mockTx) Rollback() error { return nil }

func TestDispatch_WithTopicCreatesOutboxAndAudit(t *testing.T) {
	actorUUID := uuid.MustParse("aaaaaaaa-aaaa-7aaa-aaaa-aaaaaaaaaaaa")
	ctx := ctxPck.SetActorID(context.Background(), actorUUID.String())
	tran := &mockTx{}
	mUuid := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	outboxStorer := outbox.NewMockStorer(t)
	outboxStorerTx := outbox.NewMockStorer(t)
	outboxStorer.EXPECT().
		NewWithTx(mock.AnythingOfType("*eventbus_test.mockTx")).
		Return(outboxStorerTx, nil)
	outboxStorerTx.EXPECT().
		Create(mock.Anything, mock.AnythingOfType("outbox.Outbox")).
		Return(nil)

	auditStorer := audit.NewMockStorer(t)
	auditStorerTx := audit.NewMockStorer(t)
	auditStorer.EXPECT().
		NewWithTx(mock.AnythingOfType("*eventbus_test.mockTx")).
		Return(auditStorerTx, nil)
	auditStorerTx.EXPECT().
		Create(mock.Anything, mock.AnythingOfType("audit.Audit")).
		Return(nil)

	mLogger := logger.NewMockLogger(t)
	mUuidGen := uuidgen.NewMockGenerator(t)
	mUuidGen.EXPECT().Generate().Return(mUuid, nil)
	mClock := clock.NewMockClock(t)
	mClock.EXPECT().Now().Return(mTime)

	outboxSvc := outbox_service.New(mLogger, outboxStorer, mUuidGen, mClock)
	auditSvc := audit_service.New(mLogger, auditStorer)

	dispatcher := eventbus.New(outboxSvc, auditSvc)

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

	err := dispatcher.Dispatch(ctx, tran, ev)

	assert.NoError(t, err)
}

func TestDispatch_WithoutTopicSkipsOutbox(t *testing.T) {
	actorUUID := uuid.MustParse("aaaaaaaa-aaaa-7aaa-aaaa-aaaaaaaaaaaa")
	ctx := ctxPck.SetActorID(context.Background(), actorUUID.String())
	tran := &mockTx{}

	outboxStorer := outbox.NewMockStorer(t)
	outboxStorerTx := outbox.NewMockStorer(t)
	outboxStorer.EXPECT().
		NewWithTx(mock.AnythingOfType("*eventbus_test.mockTx")).
		Return(outboxStorerTx, nil)

	auditStorer := audit.NewMockStorer(t)
	auditStorerTx := audit.NewMockStorer(t)
	auditStorer.EXPECT().
		NewWithTx(mock.AnythingOfType("*eventbus_test.mockTx")).
		Return(auditStorerTx, nil)
	auditStorerTx.EXPECT().
		Create(mock.Anything, mock.AnythingOfType("audit.Audit")).
		Return(nil)

	mLogger := logger.NewMockLogger(t)
	mUuidGen := uuidgen.NewMockGenerator(t)
	mClock := clock.NewMockClock(t)

	outboxSvc := outbox_service.New(mLogger, outboxStorer, mUuidGen, mClock)
	auditSvc := audit_service.New(mLogger, auditStorer)

	dispatcher := eventbus.New(outboxSvc, auditSvc)

	ev := event.DomainEvent{
		AggregateID: uuid.MustParse("33333333-3333-7333-3333-333333333333"),
		Payload:     map[string]string{"name": "Admin"},
		ObjEntity:   entity.New(entity.RoleEntity),
		ObjName:     name.MustParse("Admin"),
		Action:      audit.ActionCreate,
		Message:     "role CREATE",
	}

	err := dispatcher.Dispatch(ctx, tran, ev)

	assert.NoError(t, err)
}

func TestDispatch_MissingActorIDUsesZeroUUID(t *testing.T) {
	ctx := context.Background()
	tran := &mockTx{}

	outboxStorer := outbox.NewMockStorer(t)
	outboxStorerTx := outbox.NewMockStorer(t)
	outboxStorer.EXPECT().
		NewWithTx(mock.AnythingOfType("*eventbus_test.mockTx")).
		Return(outboxStorerTx, nil)

	auditStorer := audit.NewMockStorer(t)
	auditStorerTx := audit.NewMockStorer(t)
	auditStorer.EXPECT().
		NewWithTx(mock.AnythingOfType("*eventbus_test.mockTx")).
		Return(auditStorerTx, nil)
	auditStorerTx.EXPECT().
		Create(mock.Anything, mock.AnythingOfType("audit.Audit")).
		Return(nil)

	mLogger := logger.NewMockLogger(t)
	mUuidGen := uuidgen.NewMockGenerator(t)
	mClock := clock.NewMockClock(t)

	outboxSvc := outbox_service.New(mLogger, outboxStorer, mUuidGen, mClock)
	auditSvc := audit_service.New(mLogger, auditStorer)

	dispatcher := eventbus.New(outboxSvc, auditSvc)

	ev := event.DomainEvent{
		AggregateID: uuid.MustParse("44444444-4444-7444-4444-444444444444"),
		Payload:     map[string]string{"test": "data"},
		ObjEntity:   entity.New(entity.UserEntity),
		ObjName:     name.MustParse("test"),
		Action:      audit.ActionCreate,
		Message:     "test",
	}

	err := dispatcher.Dispatch(ctx, tran, ev)

	assert.NoError(t, err)
}

func TestDispatch_OutboxStorerTxError(t *testing.T) {
	ctx := context.Background()
	tran := &mockTx{}

	outboxStorer := outbox.NewMockStorer(t)
	outboxStorer.EXPECT().
		NewWithTx(mock.AnythingOfType("*eventbus_test.mockTx")).
		Return(nil, assert.AnError)

	auditStorer := audit.NewMockStorer(t)

	mLogger := logger.NewMockLogger(t)
	mUuidGen := uuidgen.NewMockGenerator(t)
	mClock := clock.NewMockClock(t)

	outboxSvc := outbox_service.New(mLogger, outboxStorer, mUuidGen, mClock)
	auditSvc := audit_service.New(mLogger, auditStorer)

	dispatcher := eventbus.New(outboxSvc, auditSvc)

	ev := event.DomainEvent{
		AggregateID: uuid.New(),
		Action:      audit.ActionCreate,
		Message:     "test",
	}

	err := dispatcher.Dispatch(ctx, tran, ev)

	assert.Error(t, err)
}

func TestDispatch_AuditStorerTxError(t *testing.T) {
	ctx := context.Background()
	tran := &mockTx{}

	outboxStorer := outbox.NewMockStorer(t)
	outboxStorerTx := outbox.NewMockStorer(t)
	outboxStorer.EXPECT().
		NewWithTx(mock.AnythingOfType("*eventbus_test.mockTx")).
		Return(outboxStorerTx, nil)

	auditStorer := audit.NewMockStorer(t)
	auditStorer.EXPECT().
		NewWithTx(mock.AnythingOfType("*eventbus_test.mockTx")).
		Return(nil, assert.AnError)

	mLogger := logger.NewMockLogger(t)
	mUuidGen := uuidgen.NewMockGenerator(t)
	mClock := clock.NewMockClock(t)

	outboxSvc := outbox_service.New(mLogger, outboxStorer, mUuidGen, mClock)
	auditSvc := audit_service.New(mLogger, auditStorer)

	dispatcher := eventbus.New(outboxSvc, auditSvc)

	ev := event.DomainEvent{
		AggregateID: uuid.New(),
		Action:      audit.ActionCreate,
		Message:     "test",
	}

	err := dispatcher.Dispatch(ctx, tran, ev)

	assert.Error(t, err)
}
