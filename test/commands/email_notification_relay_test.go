package commands_test

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/Housiadas/cerberus/internal/core/email_notification_outbox"
	"github.com/Housiadas/cerberus/internal/core/email_notification_outbox/email_notification_outbox_repo"
	"github.com/Housiadas/cerberus/internal/relay"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Housiadas/cerberus/pkg/clock"
	"github.com/Housiadas/cerberus/pkg/email"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
)

func Test_EmailNotificationRelay_RetriesFailedEntries(t *testing.T) {
	t.Parallel()

	db := sc.NewDB(t)
	ctx := context.Background()

	var buf bytes.Buffer
	traceIDFn := func(context.Context) string { return "" }
	requestIDFn := func(context.Context) string { return "" }
	log := logger.New(&buf, logger.LevelInfo, "TEST", traceIDFn, requestIDFn)

	outboxRepo := email_notification_outbox_repo.NewStore(log, db)
	uuidGen := uuidgen.NewV7()
	clk := clock.NewClock()
	outboxSvc := email_notification_outbox.NewService(log, outboxRepo, uuidGen, clk)

	payload, err := json.Marshal(map[string]string{
		"subject": "Reset your password",
		"body":    "Click here to reset your password.",
	})
	require.NoError(t, err)

	id, err := uuidGen.Generate()
	require.NoError(t, err)

	entry := email_notification_outbox.New(
		id,
		"password.reset.requested",
		"test@example.com",
		payload,
		0,
		time.Now().UTC(),
		nil,
	)

	err = outboxRepo.Create(ctx, entry)
	require.NoError(t, err)

	// Verify entry is initially queryable
	unprocessed, err := outboxSvc.QueryUnprocessed(ctx, 100, 3)
	require.NoError(t, err)
	assert.Len(t, unprocessed, 1)

	// Increment retry count to max (3)
	for i := 0; i < 3; i++ {
		err = outboxSvc.IncrementRetryCount(ctx, []uuid.UUID{id})
		require.NoError(t, err)
	}

	// Entry should no longer appear once retry_count >= maxRetries
	unprocessed, err = outboxSvc.QueryUnprocessed(ctx, 100, 3)
	require.NoError(t, err)
	assert.Len(t, unprocessed, 0, "entry should be excluded after reaching max retries")
}

func Test_EmailNotificationRelay_MarkProcessed(t *testing.T) {
	t.Parallel()

	db := sc.NewDB(t)
	ctx := context.Background()

	var buf bytes.Buffer
	traceIDFn := func(context.Context) string { return "" }
	requestIDFn := func(context.Context) string { return "" }
	log := logger.New(&buf, logger.LevelInfo, "TEST", traceIDFn, requestIDFn)

	outboxRepo := email_notification_outbox_repo.NewStore(log, db)
	uuidGen := uuidgen.NewV7()
	clk := clock.NewClock()
	outboxSvc := email_notification_outbox.NewService(log, outboxRepo, uuidGen, clk)

	payload, err := json.Marshal(map[string]string{
		"subject": "Reset your password",
		"body":    "Click here to reset your password.",
	})
	require.NoError(t, err)

	id, err := uuidGen.Generate()
	require.NoError(t, err)

	entry := email_notification_outbox.New(
		id,
		"password.reset.requested",
		"mark@example.com",
		payload,
		0,
		time.Now().UTC(),
		nil,
	)

	err = outboxRepo.Create(ctx, entry)
	require.NoError(t, err)

	// Mark as processed
	err = outboxSvc.MarkProcessed(ctx, []uuid.UUID{id})
	require.NoError(t, err)

	// Should no longer appear in unprocessed
	unprocessed, err := outboxSvc.QueryUnprocessed(ctx, 100, 3)
	require.NoError(t, err)
	assert.Len(t, unprocessed, 0, "processed entry should not appear in unprocessed query")
}

func Test_EmailNotificationRelay_ProcessesBatch_InvalidSMTP(t *testing.T) {
	t.Parallel()

	db := sc.NewDB(t)
	ctx := context.Background()

	var buf bytes.Buffer
	traceIDFn := func(context.Context) string { return "" }
	requestIDFn := func(context.Context) string { return "" }
	log := logger.New(&buf, logger.LevelInfo, "TEST", traceIDFn, requestIDFn)

	outboxRepo := email_notification_outbox_repo.NewStore(log, db)
	uuidGen := uuidgen.NewV7()
	clk := clock.NewClock()
	outboxSvc := email_notification_outbox.NewService(log, outboxRepo, uuidGen, clk)

	payload, err := json.Marshal(map[string]string{
		"subject": "Reset your password",
		"body":    "Click here to reset your password.",
	})
	require.NoError(t, err)

	id, err := uuidGen.Generate()
	require.NoError(t, err)

	entry := email_notification_outbox.New(
		id,
		"password.reset.requested",
		"fail@example.com",
		payload,
		0,
		time.Now().UTC(),
		nil,
	)

	err = outboxRepo.Create(ctx, entry)
	require.NoError(t, err)

	// Use an email client pointing to an invalid SMTP server so sends will fail
	emailClient := email.New(email.Config{
		Host:     "localhost",
		Port:     19999,
		Username: "user",
		Password: "pass",
		From:     "noreply@example.com",
	})

	relay := relay.NewEmailRelay(log, outboxSvc, emailClient, 200*time.Millisecond, 100, 3)

	relayCtx, relayCancel := context.WithCancel(ctx)
	done := make(chan struct{})

	go func() {
		relay.Start(relayCtx)
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
