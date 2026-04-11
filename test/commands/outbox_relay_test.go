package commands_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	db "github.com/Housiadas/cerberus/db/sqlc"
	"github.com/Housiadas/cerberus/internal/core/outbox"
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

	pool := sc.NewDB(t)
	store := db.NewStore(pool)

	var buf bytes.Buffer
	traceIDFn := func(context.Context) string { return "" }
	requestIDFn := func(context.Context) string { return "" }
	log := logger.New(&buf, logger.LevelInfo, "TEST", traceIDFn, requestIDFn)

	kafkaProducer := sharedKafka

	uuidGen := uuidgen.NewV7()
	clk := clock.NewClock()
	outboxSvc := outbox.NewService(log, store, uuidGen, clk)

	ctx := context.Background()

	err := outboxSvc.Create(ctx, outbox.NewOutbox{
		EventType:   "user.created",
		AggregateID: uuid.New(),
		Topic:       "user-events",
		Payload:     map[string]string{"action": "created", "name": "test-user-1"},
	})
	require.NoError(t, err)

	err = outboxSvc.Create(ctx, outbox.NewOutbox{
		EventType:   "user.updated",
		AggregateID: uuid.New(),
		Topic:       "user-events",
		Payload:     map[string]string{"action": "updated", "name": "test-user-2"},
	})
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

	pool := sc.NewDB(t)
	store := db.NewStore(pool)

	var buf bytes.Buffer
	traceIDFn := func(context.Context) string { return "" }
	requestIDFn := func(context.Context) string { return "" }
	log := logger.New(&buf, logger.LevelInfo, "TEST", traceIDFn, requestIDFn)

	uuidGen := uuidgen.NewV7()
	clk := clock.NewClock()
	outboxSvc := outbox.NewService(log, store, uuidGen, clk)

	ctx := context.Background()

	err := outboxSvc.Create(ctx, outbox.NewOutbox{
		EventType:   "user.created",
		AggregateID: uuid.New(),
		Topic:       "user-events",
		Payload:     map[string]string{"action": "created"},
	})
	require.NoError(t, err)

	// Verify entry is queryable with maxRetries=3
	unprocessed, err := outboxSvc.QueryUnprocessed(ctx, 100, 3)
	require.NoError(t, err)
	assert.Len(t, unprocessed, 1)

	// Get the ID of the created entry
	entryID := unprocessed[0].ID()

	// Increment retry count 3 times to reach the threshold
	for i := 0; i < 3; i++ {
		err = outboxSvc.IncrementRetryCount(ctx, []uuid.UUID{entryID})
		require.NoError(t, err)
	}

	// Verify entry is no longer queryable (retry_count >= maxRetries)
	unprocessed, err = outboxSvc.QueryUnprocessed(ctx, 100, 3)
	require.NoError(t, err)
	assert.Len(t, unprocessed, 0, "entry should be excluded after reaching max retries")
}

func Test_OutboxRelay_MarkProcessed(t *testing.T) {
	t.Parallel()

	pool := sc.NewDB(t)
	store := db.NewStore(pool)

	var buf bytes.Buffer
	traceIDFn := func(context.Context) string { return "" }
	requestIDFn := func(context.Context) string { return "" }
	log := logger.New(&buf, logger.LevelInfo, "TEST", traceIDFn, requestIDFn)

	uuidGen := uuidgen.NewV7()
	clk := clock.NewClock()
	outboxSvc := outbox.NewService(log, store, uuidGen, clk)

	ctx := context.Background()

	err := outboxSvc.Create(ctx, outbox.NewOutbox{
		EventType:   "user.deleted",
		AggregateID: uuid.New(),
		Topic:       "user-events",
		Payload:     map[string]string{"action": "deleted"},
	})
	require.NoError(t, err)

	// Get the ID of the created entry
	unprocessed, err := outboxSvc.QueryUnprocessed(ctx, 100, 3)
	require.NoError(t, err)
	require.Len(t, unprocessed, 1)
	entryID := unprocessed[0].ID()

	// Mark as processed
	err = outboxSvc.MarkProcessed(ctx, []uuid.UUID{entryID})
	require.NoError(t, err)

	// Verify it no longer appears in unprocessed
	unprocessed, err = outboxSvc.QueryUnprocessed(ctx, 100, 3)
	require.NoError(t, err)
	assert.Len(t, unprocessed, 0, "processed entry should not appear in unprocessed query")
}
