package cron

import (
	"context"
	"flag"
	"fmt"
	"time"

	db "github.com/Housiadas/cerberus/db/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

const retentionDays = 7

// OutboxCleanupRunner implements Runner for the outbox-cleanup cron job.
type OutboxCleanupRunner struct {
	cron *Cron
}

func NewOutboxCleanupRunner(c *Cron) *OutboxCleanupRunner {
	return &OutboxCleanupRunner{cron: c}
}

func (r *OutboxCleanupRunner) Name() string { return OutboxCleanup }
func (r *OutboxCleanupRunner) Description() string {
	return "Delete processed outbox rows older than 7 days"
}
func (r *OutboxCleanupRunner) SetupFlags(_ *flag.FlagSet) {}
func (r *OutboxCleanupRunner) Run() error {
	return r.cron.outboxCleanup()
}

// outboxCleanup deletes processed rows older than the retention period
// from both outbox and email_notification_outbox tables.
func (c *Cron) outboxCleanup() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dbPool, err := db.Open(ctx, c.db)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer dbPool.Close()

	store := db.NewStore(dbPool)

	before := pgtype.Timestamp{
		Time:  time.Now().AddDate(0, 0, -retentionDays),
		Valid: true,
	}

	outboxDeleted, err := store.DeleteProcessedOutbox(ctx, before)
	if err != nil {
		return fmt.Errorf("delete processed outbox rows: %w", err)
	}

	c.log.Info(ctx,
		"outbox-cleanup",
		"table", "outbox",
		"deleted", outboxDeleted,
	)

	notificationDeleted, err := store.DeleteProcessedNotificationOutbox(ctx, before)
	if err != nil {
		return fmt.Errorf("delete processed notification outbox rows: %w", err)
	}

	c.log.Info(
		ctx,
		"outbox-cleanup",
		"table",
		"email_notification_outbox",
		"deleted",
		notificationDeleted,
	)

	return nil
}
