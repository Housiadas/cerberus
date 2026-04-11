package commands_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	db "github.com/Housiadas/cerberus/db/sqlc"
	"github.com/Housiadas/cerberus/internal/core/email_notification_outbox"
	"github.com/Housiadas/cerberus/internal/sdk/relay"
	"github.com/Housiadas/cerberus/pkg/clock"
	"github.com/Housiadas/cerberus/pkg/email"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_EmailNotificationRelay_RetriesFailedEntries(t *testing.T) {
	t.Parallel()

	pool := sc.NewDB(t)
	store := db.NewStore(pool)
	ctx := context.Background()

	var buf bytes.Buffer
	traceIDFn := func(context.Context) string { return "" }
	requestIDFn := func(context.Context) string { return "" }
	log := logger.New(&buf, logger.LevelInfo, "TEST", traceIDFn, requestIDFn)

	uuidGen := uuidgen.NewV7()
	clk := clock.NewClock()
	outboxSvc := email_notification_outbox.NewService(log, store, uuidGen, clk)

	err := outboxSvc.Create(ctx, email_notification_outbox.NewEmailNotificationOutbox{
		EventType: "password.reset.requested",
		ToEmail:   "test@example.com",
		Payload: map[string]string{
			"subject": "Reset your password",
			"body":    "Click here to reset your password.",
		},
	})
	require.NoError(t, err)

	// Verify entry is initially queryable
	unprocessed, err := outboxSvc.QueryUnprocessed(ctx, 100, 3)
	require.NoError(t, err)
	assert.Len(t, unprocessed, 1)

	// Get the ID of the created entry
	entryID := unprocessed[0].ID()

	// Increment retry count to max (3)
	for i := 0; i < 3; i++ {
		err = outboxSvc.IncrementRetryCount(ctx, []uuid.UUID{entryID})
		require.NoError(t, err)
	}

	// Entry should no longer appear once retry_count >= maxRetries
	unprocessed, err = outboxSvc.QueryUnprocessed(ctx, 100, 3)
	require.NoError(t, err)
	assert.Len(t, unprocessed, 0, "entry should be excluded after reaching max retries")
}

func Test_EmailNotificationRelay_MarkProcessed(t *testing.T) {
	t.Parallel()

	pool := sc.NewDB(t)
	store := db.NewStore(pool)
	ctx := context.Background()

	var buf bytes.Buffer
	traceIDFn := func(context.Context) string { return "" }
	requestIDFn := func(context.Context) string { return "" }
	log := logger.New(&buf, logger.LevelInfo, "TEST", traceIDFn, requestIDFn)

	uuidGen := uuidgen.NewV7()
	clk := clock.NewClock()
	outboxSvc := email_notification_outbox.NewService(log, store, uuidGen, clk)

	err := outboxSvc.Create(ctx, email_notification_outbox.NewEmailNotificationOutbox{
		EventType: "password.reset.requested",
		ToEmail:   "mark@example.com",
		Payload: map[string]string{
			"subject": "Reset your password",
			"body":    "Click here to reset your password.",
		},
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

	// Should no longer appear in unprocessed
	unprocessed, err = outboxSvc.QueryUnprocessed(ctx, 100, 3)
	require.NoError(t, err)
	assert.Len(t, unprocessed, 0, "processed entry should not appear in unprocessed query")
}

func Test_EmailNotificationRelay_ProcessesBatch_InvalidSMTP(t *testing.T) {
	t.Parallel()

	pool := sc.NewDB(t)
	store := db.NewStore(pool)
	ctx := context.Background()

	var buf bytes.Buffer
	traceIDFn := func(context.Context) string { return "" }
	requestIDFn := func(context.Context) string { return "" }
	log := logger.New(&buf, logger.LevelInfo, "TEST", traceIDFn, requestIDFn)

	uuidGen := uuidgen.NewV7()
	clk := clock.NewClock()
	outboxSvc := email_notification_outbox.NewService(log, store, uuidGen, clk)

	err := outboxSvc.Create(ctx, email_notification_outbox.NewEmailNotificationOutbox{
		EventType: "password.reset.requested",
		ToEmail:   "fail@example.com",
		Payload: map[string]string{
			"subject": "Reset your password",
			"body":    "Click here to reset your password.",
		},
	})
	require.NoError(t, err)

	// Use an email client pointing to an invalid SMTP server so sends will fail
	emailClient := email.New(email.Config{
		Host:     "localhost",
		Port:     19999,
		Username: "user",
		Password: "pass",
		From:     "noreply@example.com",
	})

	emailRelay := relay.NewEmailRelay(log, outboxSvc, emailClient, 200*time.Millisecond, 100, 3)

	relayCtx, relayCancel := context.WithCancel(ctx)
	done := make(chan struct{})

	go func() {
		emailRelay.Start(relayCtx)
		close(done)
	}()

	// Wait for relay to attempt sending and increment retry count
	assert.Eventually(t, func() bool {
		unprocessed, queryErr := outboxSvc.QueryUnprocessed(ctx, 100, 1)
		if queryErr != nil {
			return false
		}
		// After 1 failed attempt, retry_count == 1, which means it won't appear when maxRetries=1
		return len(unprocessed) == 0
	}, 10*time.Second, 300*time.Millisecond, "relay should increment retry count after failed send")

	relayCancel()
	<-done
}
