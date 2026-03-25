package command

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Housiadas/cerberus/internal/core/email_notification_outbox/email_notification_outbox_repo"
	"github.com/Housiadas/cerberus/internal/core/email_notification_outbox/email_notification_outbox_service"
	"github.com/Housiadas/cerberus/internal/relay"
	"github.com/Housiadas/cerberus/pkg/clock"
	"github.com/Housiadas/cerberus/pkg/email"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/Housiadas/cerberus/pkg/telemetry"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
)

const (
	emailBatchSize  = 100
	emailMaxRetries = 3
)

// EmailNotificationRelay starts the email notification relay process that polls
// the email_notification_outbox table and sends emails via SMTP.
func (cmd *Command) EmailNotificationRelay() error {
	db, err := pgsql.Open(cmd.db)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer db.Close()

	emailClient := email.New(cmd.emailConfig)

	emailOutboxRepo := email_notification_outbox_repo.NewStore(cmd.log, db)
	emailOutboxSvc := email_notification_outbox_service.New(
		cmd.log, emailOutboxRepo, uuidgen.NewV7(), clock.NewClock(),
	)

	emailRelay := relay.NewEmailRelay(
		cmd.log,
		emailOutboxSvc,
		emailClient,
		3*time.Second,
		emailBatchSize,
		emailMaxRetries,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctx = telemetry.InjectTracing(ctx, cmd.tracer)

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	done := make(chan struct{})

	go func() {
		emailRelay.Start(ctx)
		close(done)
	}()

	cmd.log.Info(ctx, "startup", "status", "email notification emailRelay started")

	sig := <-shutdown
	cmd.log.Info(ctx, "shutdown", "status", "email notification emailRelay stopping", "signal", sig)

	cancel()
	<-done

	cmd.log.Info(ctx, "shutdown", "status", "email notification emailRelay stopped")

	return nil
}
