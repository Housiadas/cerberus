package outbox_relay_test

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/Housiadas/cerberus/internal/app/relay"
	"github.com/Housiadas/cerberus/internal/app/repo/outbox_repo"
	"github.com/Housiadas/cerberus/internal/core/domain/outbox"
	"github.com/Housiadas/cerberus/internal/core/service/outbox_service"
	"github.com/Housiadas/cerberus/internal/utils/dbtest"
	"github.com/Housiadas/cerberus/internal/utils/kafkatest"
	"github.com/Housiadas/cerberus/pkg/clock"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_OutboxRelay_ProcessesEntries(t *testing.T) {
	t.Parallel()

	db := dbtest.New(t, "Test_OutboxRelay_ProcessesEntries")

	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", "", "")

	kafkaProducer := kafkatest.New(t)

	outboxRepo := outbox_repo.NewStore(log, db)
	uuidGen := uuidgen.NewV7()
	clk := clock.NewClock()
	outboxSvc := outbox_service.New(log, outboxRepo, uuidGen, clk)

	// Insert test outbox entries directly via the service
	ctx := context.Background()

	aggregateID1 := uuid.New()
	aggregateID2 := uuid.New()

	payload1, err := json.Marshal(map[string]string{"action": "created", "name": "test-user-1"})
	require.NoError(t, err)

	payload2, err := json.Marshal(map[string]string{"action": "updated", "name": "test-user-2"})
	require.NoError(t, err)

	// Insert entries directly into DB via repo
	now := time.Now().UTC()
	id1, err := uuidGen.Generate()
	require.NoError(t, err)
	id2, err := uuidGen.Generate()
	require.NoError(t, err)

	entry1 := outbox.Outbox{
		ID:          id1,
		EventType:   "user.created",
		AggregateID: aggregateID1,
		Topic:       "user-events",
		Payload:     payload1,
		RetryCount:  0,
		CreatedAt:   now,
	}

	entry2 := outbox.Outbox{
		ID:          id2,
		EventType:   "user.updated",
		AggregateID: aggregateID2,
		Topic:       "user-events",
		Payload:     payload2,
		RetryCount:  0,
		CreatedAt:   now,
	}

	err = outboxRepo.Create(ctx, entry1)
	require.NoError(t, err)

	err = outboxRepo.Create(ctx, entry2)
	require.NoError(t, err)

	// Verify entries are unprocessed
	unprocessed, err := outboxSvc.QueryUnprocessed(ctx, 100, 3)
	require.NoError(t, err)
	assert.Len(t, unprocessed, 2)

	// Start the relay with a short interval
	outboxRelay := relay.New(log, outboxSvc, kafkaProducer, 500*time.Millisecond, 100, 3)

	relayCtx, relayCancel := context.WithCancel(ctx)

	done := make(chan struct{})
	go func() {
		outboxRelay.Start(relayCtx)
		close(done)
	}()

	// Wait for the relay to process the entries
	assert.Eventually(t, func() bool {
		entries, err := outboxSvc.QueryUnprocessed(ctx, 100, 3)
		if err != nil {
			return false
		}
		return len(entries) == 0
	}, 10*time.Second, 500*time.Millisecond, "outbox entries should be processed")

	relayCancel()
	<-done
}

func Test_OutboxRelay_RetriesFailedEntries(t *testing.T) {
	t.Parallel()

	db := dbtest.New(t, "Test_OutboxRelay_RetriesFailedEntries")

	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", "", "")

	outboxRepo := outbox_repo.NewStore(log, db)
	uuidGen := uuidgen.NewV7()
	clk := clock.NewClock()
	outboxSvc := outbox_service.New(log, outboxRepo, uuidGen, clk)

	ctx := context.Background()

	id1, err := uuidGen.Generate()
	require.NoError(t, err)

	payload, err := json.Marshal(map[string]string{"action": "created"})
	require.NoError(t, err)

	entry := outbox.Outbox{
		ID:          id1,
		EventType:   "user.created",
		AggregateID: uuid.New(),
		Topic:       "user-events",
		Payload:     payload,
		RetryCount:  0,
		CreatedAt:   time.Now().UTC(),
	}

	err = outboxRepo.Create(ctx, entry)
	require.NoError(t, err)

	// Verify entry is queryable with maxRetries=3
	unprocessed, err := outboxSvc.QueryUnprocessed(ctx, 100, 3)
	require.NoError(t, err)
	assert.Len(t, unprocessed, 1)

	// Increment retry count 3 times to reach the threshold
	for i := 0; i < 3; i++ {
		err = outboxSvc.IncrementRetryCount(ctx, []uuid.UUID{id1})
		require.NoError(t, err)
	}

	// Verify entry is no longer queryable (retry_count >= maxRetries)
	unprocessed, err = outboxSvc.QueryUnprocessed(ctx, 100, 3)
	require.NoError(t, err)
	assert.Len(t, unprocessed, 0, "entry should be excluded after reaching max retries")
}

func Test_OutboxRelay_MarkProcessed(t *testing.T) {
	t.Parallel()

	db := dbtest.New(t, "Test_OutboxRelay_MarkProcessed")

	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", "", "")

	outboxRepo := outbox_repo.NewStore(log, db)
	uuidGen := uuidgen.NewV7()
	clk := clock.NewClock()
	outboxSvc := outbox_service.New(log, outboxRepo, uuidGen, clk)

	ctx := context.Background()

	id1, err := uuidGen.Generate()
	require.NoError(t, err)

	payload, err := json.Marshal(map[string]string{"action": "deleted"})
	require.NoError(t, err)

	entry := outbox.Outbox{
		ID:          id1,
		EventType:   "user.deleted",
		AggregateID: uuid.New(),
		Topic:       "user-events",
		Payload:     payload,
		RetryCount:  0,
		CreatedAt:   time.Now().UTC(),
	}

	err = outboxRepo.Create(ctx, entry)
	require.NoError(t, err)

	// Mark as processed
	err = outboxSvc.MarkProcessed(ctx, []uuid.UUID{id1})
	require.NoError(t, err)

	// Verify it no longer appears in unprocessed
	unprocessed, err := outboxSvc.QueryUnprocessed(ctx, 100, 3)
	require.NoError(t, err)
	assert.Len(t, unprocessed, 0, "processed entry should not appear in unprocessed query")
}
