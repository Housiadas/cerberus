package email_notification_outbox_service_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Housiadas/cerberus/internal/core/domain/email_notification_outbox"
	"github.com/Housiadas/cerberus/internal/core/service/email_notification_outbox_service"
	"github.com/Housiadas/cerberus/pkg/clock"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
)

func TestService_Create_Successful(t *testing.T) {
	ctx := context.Background()
	id := uuid.MustParse("11111111-1111-7111-1111-111111111111")
	now := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)

	payload := map[string]string{"resetUrl": "http://localhost:3000/reset?token=abc"}
	payloadBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	no := email_notification_outbox.NewEmailNotificationOutbox{
		EventType: "password.reset.requested",
		ToEmail:   "john@example.com",
		Payload:   payload,
	}

	expectedEntry := email_notification_outbox.New(id, no.EventType, no.ToEmail, payloadBytes, 0, now, nil)

	mLogger := logger.NewMockLogger(t)
	mUuidGen := uuidgen.NewMockGenerator(t)
	mUuidGen.EXPECT().Generate().Return(id, nil)

	mClock := clock.NewMockClock(t)
	mClock.EXPECT().Now().Return(now)

	mStorer := email_notification_outbox.NewMockStorer(t)
	mStorer.EXPECT().Create(ctx, expectedEntry).Return(nil)

	sut := email_notification_outbox_service.New(mLogger, mStorer, mUuidGen, mClock)
	err = sut.Create(ctx, no)

	assert.NoError(t, err)
}

func TestService_Create_UUIDError(t *testing.T) {
	ctx := context.Background()

	no := email_notification_outbox.NewEmailNotificationOutbox{
		EventType: "password.reset.requested",
		ToEmail:   "john@example.com",
		Payload:   map[string]string{"key": "value"},
	}

	mLogger := logger.NewMockLogger(t)
	mUuidGen := uuidgen.NewMockGenerator(t)
	mUuidGen.EXPECT().Generate().Return(uuid.UUID{}, errors.New("uuid error"))

	mClock := clock.NewMockClock(t)
	mStorer := email_notification_outbox.NewMockStorer(t)

	sut := email_notification_outbox_service.New(mLogger, mStorer, mUuidGen, mClock)
	err := sut.Create(ctx, no)

	assert.Error(t, err)
}

func TestService_Create_StorerError(t *testing.T) {
	ctx := context.Background()
	id := uuid.MustParse("11111111-1111-7111-1111-111111111111")
	now := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)

	payload := map[string]string{"key": "value"}
	payloadBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	no := email_notification_outbox.NewEmailNotificationOutbox{
		EventType: "password.reset.requested",
		ToEmail:   "john@example.com",
		Payload:   payload,
	}

	expectedEntry := email_notification_outbox.New(id, no.EventType, no.ToEmail, payloadBytes, 0, now, nil)

	mLogger := logger.NewMockLogger(t)
	mUuidGen := uuidgen.NewMockGenerator(t)
	mUuidGen.EXPECT().Generate().Return(id, nil)

	mClock := clock.NewMockClock(t)
	mClock.EXPECT().Now().Return(now)

	mStorer := email_notification_outbox.NewMockStorer(t)
	mStorer.EXPECT().Create(ctx, expectedEntry).Return(errors.New("db error"))

	sut := email_notification_outbox_service.New(mLogger, mStorer, mUuidGen, mClock)
	err = sut.Create(ctx, no)

	assert.Error(t, err)
}

func TestService_QueryUnprocessed_Successful(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)

	entries := []email_notification_outbox.EmailNotificationOutbox{
		email_notification_outbox.New(uuid.New(), "type1", "a@example.com", []byte(`{}`), 0, now, nil),
		email_notification_outbox.New(uuid.New(), "type2", "b@example.com", []byte(`{}`), 1, now, nil),
	}

	mLogger := logger.NewMockLogger(t)
	mUuidGen := uuidgen.NewMockGenerator(t)
	mClock := clock.NewMockClock(t)
	mStorer := email_notification_outbox.NewMockStorer(t)
	mStorer.EXPECT().QueryUnprocessed(ctx, 10, 3).Return(entries, nil)

	sut := email_notification_outbox_service.New(mLogger, mStorer, mUuidGen, mClock)
	got, err := sut.QueryUnprocessed(ctx, 10, 3)

	require.NoError(t, err)
	assert.Len(t, got, 2)
}

func TestService_QueryUnprocessed_Error(t *testing.T) {
	ctx := context.Background()

	mLogger := logger.NewMockLogger(t)
	mUuidGen := uuidgen.NewMockGenerator(t)
	mClock := clock.NewMockClock(t)
	mStorer := email_notification_outbox.NewMockStorer(t)
	mStorer.EXPECT().QueryUnprocessed(ctx, 10, 3).Return(nil, errors.New("db error"))

	sut := email_notification_outbox_service.New(mLogger, mStorer, mUuidGen, mClock)
	_, err := sut.QueryUnprocessed(ctx, 10, 3)

	assert.Error(t, err)
}

func TestService_IncrementRetryCount_Successful(t *testing.T) {
	ctx := context.Background()
	ids := []uuid.UUID{uuid.New(), uuid.New()}

	mLogger := logger.NewMockLogger(t)
	mUuidGen := uuidgen.NewMockGenerator(t)
	mClock := clock.NewMockClock(t)
	mStorer := email_notification_outbox.NewMockStorer(t)
	mStorer.EXPECT().IncrementRetryCount(ctx, ids).Return(nil)

	sut := email_notification_outbox_service.New(mLogger, mStorer, mUuidGen, mClock)
	err := sut.IncrementRetryCount(ctx, ids)

	assert.NoError(t, err)
}

func TestService_IncrementRetryCount_Error(t *testing.T) {
	ctx := context.Background()
	ids := []uuid.UUID{uuid.New()}

	mLogger := logger.NewMockLogger(t)
	mUuidGen := uuidgen.NewMockGenerator(t)
	mClock := clock.NewMockClock(t)
	mStorer := email_notification_outbox.NewMockStorer(t)
	mStorer.EXPECT().IncrementRetryCount(ctx, ids).Return(errors.New("db error"))

	sut := email_notification_outbox_service.New(mLogger, mStorer, mUuidGen, mClock)
	err := sut.IncrementRetryCount(ctx, ids)

	assert.Error(t, err)
}

func TestService_MarkProcessed_Successful(t *testing.T) {
	ctx := context.Background()
	ids := []uuid.UUID{uuid.New()}
	now := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)

	mLogger := logger.NewMockLogger(t)
	mUuidGen := uuidgen.NewMockGenerator(t)
	mClock := clock.NewMockClock(t)
	mClock.EXPECT().Now().Return(now)

	mStorer := email_notification_outbox.NewMockStorer(t)
	mStorer.EXPECT().MarkProcessed(ctx, ids, now).Return(nil)

	sut := email_notification_outbox_service.New(mLogger, mStorer, mUuidGen, mClock)
	err := sut.MarkProcessed(ctx, ids)

	assert.NoError(t, err)
}

func TestService_MarkProcessed_Error(t *testing.T) {
	ctx := context.Background()
	ids := []uuid.UUID{uuid.New()}
	now := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)

	mLogger := logger.NewMockLogger(t)
	mUuidGen := uuidgen.NewMockGenerator(t)
	mClock := clock.NewMockClock(t)
	mClock.EXPECT().Now().Return(now)

	mStorer := email_notification_outbox.NewMockStorer(t)
	mStorer.EXPECT().MarkProcessed(ctx, ids, now).Return(errors.New("db error"))

	sut := email_notification_outbox_service.New(mLogger, mStorer, mUuidGen, mClock)
	err := sut.MarkProcessed(ctx, ids)

	assert.Error(t, err)
}
