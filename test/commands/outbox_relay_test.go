package commands_test

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/Housiadas/cerberus/internal/core/outbox"
	"github.com/Housiadas/cerberus/internal/core/outbox/outbox_repo"
	"github.com/Housiadas/cerberus/internal/sdk/relay"
	"github.com/Housiadas/cerberus/pkg/clock"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_OutboxRelay_ProcessesEntries(t *testing.T) {
	t.Parallel()

	db := sc.NewDB(t)

	var buf bytes.Buffer
	traceIDFn := func(context.Context) string {
		return ""
	}
	requestIDFn := func(context.Context) string {
		return ""
	}
	log := logger.New(&buf, logger.LevelInfo, "TEST", traceIDFn, requestIDFn)

	kafkaProducer := sharedKafka

	outboxRepo := outbox_repo.NewStore(log, db)
	uuidGen := uuidgen.NewV7()
	clk := clock.NewClock()
	outboxSvc := outbox.NewService(log, outboxRepo, uuidGen, clk)

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

	entry1 := outbox.New(id1, "user.created", aggregateID1, "user-events", payload1, 0, now, nil)
	entry2 := outbox.New(id2, "user.updated", aggregateID2, "user-events", payload2, 0, now, nil)

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

	db := sc.NewDB(t)

	var buf bytes.Buffer
	traceIDFn := func(context.Context) string {
		return ""
	}
	requestIDFn := func(context.Context) string {
		return ""
	}
	log := logger.New(&buf, logger.LevelInfo, "TEST", traceIDFn, requestIDFn)

	outboxRepo := outbox_repo.NewStore(log, db)
	uuidGen := uuidgen.NewV7()
	clk := clock.NewClock()
	outboxSvc := outbox.NewService(log, outboxRepo, uuidGen, clk)

	ctx := context.Background()

	id1, err := uuidGen.Generate()
	require.NoError(t, err)

	payload, err := json.Marshal(map[string]string{"action": "created"})
	require.NoError(t, err)

	entry := outbox.New(id1, "user.created", uuid.New(), "user-events", payload, 0, time.Now().UTC(), nil)

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

	db := sc.NewDB(t)

	var buf bytes.Buffer
	traceIDFn := func(context.Context) string {
		return ""
	}
	requestIDFn := func(context.Context) string {
		return ""
	}
	log := logger.New(&buf, logger.LevelInfo, "TEST", traceIDFn, requestIDFn)

	outboxRepo := outbox_repo.NewStore(log, db)
	uuidGen := uuidgen.NewV7()
	clk := clock.NewClock()
	outboxSvc := outbox.NewService(log, outboxRepo, uuidGen, clk)

	ctx := context.Background()

	id1, err := uuidGen.Generate()
	require.NoError(t, err)

	payload, err := json.Marshal(map[string]string{"action": "deleted"})
	require.NoError(t, err)

	entry := outbox.New(id1, "user.deleted", uuid.New(), "user-events", payload, 0, time.Now().UTC(), nil)

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
