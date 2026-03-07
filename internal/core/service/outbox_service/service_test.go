package outbox_service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Housiadas/cerberus/internal/core/domain/outbox"
	"github.com/Housiadas/cerberus/internal/core/service/outbox_service"
	"github.com/Housiadas/cerberus/pkg/clock"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
)

func TestService_Create_Successful(t *testing.T) {
	ctx := context.Background()
	mUuid := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	no := outbox.NewOutbox{
		EventType:   "user.created",
		AggregateID: uuid.MustParse("22222222-2222-7222-2222-222222222222"),
		Topic:       "user-events",
		Payload:     map[string]string{"name": "John"},
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := outbox.NewMockStorer(t)
	mStorer.EXPECT().Create(ctx, mock.AnythingOfType("outbox.Outbox")).Return(nil)

	mUuidGen := uuidgen.NewMockGenerator(t)
	mUuidGen.EXPECT().Generate().Return(mUuid, nil)

	mClock := clock.NewMockClock(t)
	mClock.EXPECT().Now().Return(mTime)

	sut := outbox_service.New(mLogger, mStorer, mUuidGen, mClock)
	err := sut.Create(ctx, no)

	assert.NoError(t, err)
}

func TestService_Create_UuidError(t *testing.T) {
	ctx := context.Background()

	no := outbox.NewOutbox{
		EventType:   "user.created",
		AggregateID: uuid.MustParse("22222222-2222-7222-2222-222222222222"),
		Topic:       "user-events",
		Payload:     map[string]string{"name": "John"},
	}

	mLogger := logger.NewMockLogger(t)
	mStorer := outbox.NewMockStorer(t)

	mUuidGen := uuidgen.NewMockGenerator(t)
	mUuidGen.EXPECT().Generate().Return(uuid.Nil, errors.New("uuid error"))

	mClock := clock.NewMockClock(t)

	sut := outbox_service.New(mLogger, mStorer, mUuidGen, mClock)
	err := sut.Create(ctx, no)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "uuid error")
}

func TestService_Create_MarshalError(t *testing.T) {
	ctx := context.Background()
	mUuid := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")

	no := outbox.NewOutbox{
		EventType:   "user.created",
		AggregateID: uuid.MustParse("22222222-2222-7222-2222-222222222222"),
		Topic:       "user-events",
		Payload:     make(chan int), // channels cannot be marshalled
	}

	mLogger := logger.NewMockLogger(t)
	mStorer := outbox.NewMockStorer(t)

	mUuidGen := uuidgen.NewMockGenerator(t)
	mUuidGen.EXPECT().Generate().Return(mUuid, nil)

	mClock := clock.NewMockClock(t)

	sut := outbox_service.New(mLogger, mStorer, mUuidGen, mClock)
	err := sut.Create(ctx, no)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "marshal payload error")
}

func TestService_Create_StorerError(t *testing.T) {
	ctx := context.Background()
	mUuid := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	no := outbox.NewOutbox{
		EventType:   "user.created",
		AggregateID: uuid.MustParse("22222222-2222-7222-2222-222222222222"),
		Topic:       "user-events",
		Payload:     map[string]string{"name": "John"},
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := outbox.NewMockStorer(t)
	mStorer.EXPECT().Create(ctx, mock.AnythingOfType("outbox.Outbox")).Return(errors.New("db error"))

	mUuidGen := uuidgen.NewMockGenerator(t)
	mUuidGen.EXPECT().Generate().Return(mUuid, nil)

	mClock := clock.NewMockClock(t)
	mClock.EXPECT().Now().Return(mTime)

	sut := outbox_service.New(mLogger, mStorer, mUuidGen, mClock)
	err := sut.Create(ctx, no)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}

func TestService_QueryUnprocessed_Successful(t *testing.T) {
	ctx := context.Background()
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	expected := []outbox.Outbox{
		outbox.New(
			uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
			"user.created",
			uuid.MustParse("22222222-2222-7222-2222-222222222222"),
			"user-events",
			[]byte(`{"name":"John"}`),
			0,
			mTime,
			nil,
		),
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := outbox.NewMockStorer(t)
	mStorer.EXPECT().QueryUnprocessed(ctx, 10, 5).Return(expected, nil)

	mUuidGen := uuidgen.NewMockGenerator(t)
	mClock := clock.NewMockClock(t)

	sut := outbox_service.New(mLogger, mStorer, mUuidGen, mClock)
	entries, err := sut.QueryUnprocessed(ctx, 10, 5)

	assert.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, expected[0].ID(), entries[0].ID())
}

func TestService_QueryUnprocessed_Error(t *testing.T) {
	ctx := context.Background()

	mLogger := logger.NewMockLogger(t)

	mStorer := outbox.NewMockStorer(t)
	mStorer.EXPECT().QueryUnprocessed(ctx, 10, 5).Return(nil, errors.New("query error"))

	mUuidGen := uuidgen.NewMockGenerator(t)
	mClock := clock.NewMockClock(t)

	sut := outbox_service.New(mLogger, mStorer, mUuidGen, mClock)
	_, err := sut.QueryUnprocessed(ctx, 10, 5)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "query error")
}

func TestService_MarkProcessed_Successful(t *testing.T) {
	ctx := context.Background()
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)
	ids := []uuid.UUID{uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")}

	mLogger := logger.NewMockLogger(t)

	mStorer := outbox.NewMockStorer(t)
	mStorer.EXPECT().MarkProcessed(ctx, ids, mTime).Return(nil)

	mUuidGen := uuidgen.NewMockGenerator(t)

	mClock := clock.NewMockClock(t)
	mClock.EXPECT().Now().Return(mTime)

	sut := outbox_service.New(mLogger, mStorer, mUuidGen, mClock)
	err := sut.MarkProcessed(ctx, ids)

	assert.NoError(t, err)
}

func TestService_MarkProcessed_Error(t *testing.T) {
	ctx := context.Background()
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)
	ids := []uuid.UUID{uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")}

	mLogger := logger.NewMockLogger(t)

	mStorer := outbox.NewMockStorer(t)
	mStorer.EXPECT().MarkProcessed(ctx, ids, mTime).Return(errors.New("mark error"))

	mUuidGen := uuidgen.NewMockGenerator(t)

	mClock := clock.NewMockClock(t)
	mClock.EXPECT().Now().Return(mTime)

	sut := outbox_service.New(mLogger, mStorer, mUuidGen, mClock)
	err := sut.MarkProcessed(ctx, ids)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mark error")
}
