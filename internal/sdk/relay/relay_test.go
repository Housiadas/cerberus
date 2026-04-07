package relay_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/Housiadas/cerberus/internal/core/outbox"
	"github.com/Housiadas/cerberus/internal/sdk/relay"
	loggermocks "github.com/Housiadas/cerberus/pkg/logger/mocks"
	ckafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRelay_ProcessBatch_Successful(t *testing.T) {
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)
	id1 := uuid.MustParse("11111111-1111-7111-1111-111111111111")
	id2 := uuid.MustParse("22222222-2222-7222-2222-222222222222")

	entries := []outbox.Outbox{
		outbox.New(id1, "user.created", uuid.MustParse("aaaaaaaa-aaaa-7aaa-aaaa-aaaaaaaaaaaa"), "user-events", json.RawMessage(`{"name":"John"}`), 0, mTime, nil),
		outbox.New(id2, "user.updated", uuid.MustParse("bbbbbbbb-bbbb-7bbb-bbbb-bbbbbbbbbbbb"), "user-events", json.RawMessage(`{"name":"Jane"}`), 0, mTime, nil),
	}

	mLogger := loggermocks.NewMockLogger(t)
	mLogger.EXPECT().Info(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Maybe()
	mLogger.EXPECT().Info(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()

	mStorer := newMockoutboxer(t)
	mStorer.EXPECT().QueryUnprocessed(mock.Anything, 100, 5).
		Return(entries, nil).Once()
	mStorer.EXPECT().QueryUnprocessed(mock.Anything, 100, 5).
		Return(nil, nil).Maybe()
	mStorer.EXPECT().MarkProcessed(mock.Anything, []uuid.UUID{id1, id2}).Return(nil).Once()

	mProducer := newMockproducer(t)
	mProducer.EXPECT().Produce(mock.Anything, mock.AnythingOfType("*kafka.Message")).Return(nil).Times(2)
	mProducer.EXPECT().Flush(mock.AnythingOfType("int")).Return(0).Maybe()

	r := relay.New(mLogger, mStorer, mProducer, 50*time.Millisecond, 100, 5)

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	r.Start(ctx)

	// Assertions are validated by mock expectations
}

func TestRelay_ProcessBatch_PartialFailure(t *testing.T) {
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)
	id1 := uuid.MustParse("11111111-1111-7111-1111-111111111111")
	id2 := uuid.MustParse("22222222-2222-7222-2222-222222222222")

	entries := []outbox.Outbox{
		outbox.New(id1, "user.created", uuid.MustParse("aaaaaaaa-aaaa-7aaa-aaaa-aaaaaaaaaaaa"), "user-events", json.RawMessage(`{"name":"John"}`), 0, mTime, nil),
		outbox.New(id2, "user.updated", uuid.MustParse("bbbbbbbb-bbbb-7bbb-bbbb-bbbbbbbbbbbb"), "user-events", json.RawMessage(`{"name":"Jane"}`), 0, mTime, nil),
	}

	mLogger := loggermocks.NewMockLogger(t)
	mLogger.EXPECT().Info(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()
	mLogger.EXPECT().Info(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()
	mLogger.EXPECT().Error(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()

	mStorer := newMockoutboxer(t)
	mStorer.EXPECT().QueryUnprocessed(mock.Anything, 100, 5).Return(entries, nil).Once()
	mStorer.EXPECT().QueryUnprocessed(mock.Anything, 100, 5).Return(nil, nil).Maybe()
	// Only id2 should be marked processed since id1 fails
	mStorer.EXPECT().MarkProcessed(mock.Anything, []uuid.UUID{id2}).Return(nil).Once()
	mStorer.EXPECT().IncrementRetryCount(mock.Anything, []uuid.UUID{id1}).Return(nil).Once()

	mProducer := newMockproducer(t)
	// The first call fails, the second succeeds
	mProducer.EXPECT().Produce(mock.Anything, mock.MatchedBy(func(msg *ckafka.Message) bool {
		return string(msg.Key) == uuid.MustParse("aaaaaaaa-aaaa-7aaa-aaaa-aaaaaaaaaaaa").String()
	})).Return(errors.New("produce error")).Once()
	mProducer.EXPECT().Produce(mock.Anything, mock.MatchedBy(func(msg *ckafka.Message) bool {
		return string(msg.Key) == uuid.MustParse("bbbbbbbb-bbbb-7bbb-bbbb-bbbbbbbbbbbb").String()
	})).Return(nil).Once()
	mProducer.EXPECT().Flush(mock.AnythingOfType("int")).Return(0).Maybe()

	r := relay.New(mLogger, mStorer, mProducer, 50*time.Millisecond, 100, 5)

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	r.Start(ctx)
}

func TestRelay_ProcessBatch_Empty(t *testing.T) {
	mLogger := loggermocks.NewMockLogger(t)
	mLogger.EXPECT().Info(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()
	mLogger.EXPECT().Info(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()

	mStorer := newMockoutboxer(t)
	mStorer.EXPECT().QueryUnprocessed(mock.Anything, 100, 5).Return(nil, nil).Maybe()

	mProducer := newMockproducer(t)
	mProducer.EXPECT().Flush(mock.AnythingOfType("int")).Return(0).Maybe()

	r := relay.New(mLogger, mStorer, mProducer, 50*time.Millisecond, 100, 5)

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	r.Start(ctx)

	// No produce calls expected - validated by mock
	assert.True(t, true)
}

func TestRelay_ProcessBatch_AllFailed_DeadLetter(t *testing.T) {
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)
	id1 := uuid.MustParse("11111111-1111-7111-1111-111111111111")

	entries := []outbox.Outbox{
		outbox.New(id1, "user.created", uuid.MustParse("aaaaaaaa-aaaa-7aaa-aaaa-aaaaaaaaaaaa"), "user-events", json.RawMessage(`{"name":"John"}`), 0, mTime, nil),
	}

	mLogger := loggermocks.NewMockLogger(t)
	mLogger.EXPECT().Info(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()
	mLogger.EXPECT().Info(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()
	mLogger.EXPECT().Error(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()

	mStorer := newMockoutboxer(t)
	mStorer.EXPECT().QueryUnprocessed(mock.Anything, 100, 5).Return(entries, nil).Once()
	mStorer.EXPECT().QueryUnprocessed(mock.Anything, 100, 5).Return(nil, nil).Maybe()
	mStorer.EXPECT().IncrementRetryCount(mock.Anything, []uuid.UUID{id1}).Return(nil).Once()

	mProducer := newMockproducer(t)
	mProducer.EXPECT().Produce(mock.Anything, mock.AnythingOfType("*kafka.Message")).Return(errors.New("produce error")).Once()
	mProducer.EXPECT().Flush(mock.AnythingOfType("int")).Return(0).Maybe()

	r := relay.New(mLogger, mStorer, mProducer, 50*time.Millisecond, 100, 5)

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	r.Start(ctx)

	// MarkProcessed should NOT be called — validated by mock expectations
}
