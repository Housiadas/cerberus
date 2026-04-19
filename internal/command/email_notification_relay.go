package command

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	db "github.com/Housiadas/cerberus/db/sqlc"
	"github.com/Housiadas/cerberus/internal/core/email_notification_outbox"
	"github.com/Housiadas/cerberus/internal/sdk/relay"
	"github.com/Housiadas/cerberus/pkg/clock"
	"github.com/Housiadas/cerberus/pkg/email"
	"github.com/Housiadas/cerberus/pkg/telemetry"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
)

const (
	emailBatchSize  = 100
	emailMaxRetries = 3
)

// EmailNotificationRelayRunner implements Runner for the email-notification-relay command.
type EmailNotificationRelayRunner struct {
	cmd *Command
}

func NewEmailNotificationRelayRunner(cmd *Command) *EmailNotificationRelayRunner {
	return &EmailNotificationRelayRunner{cmd: cmd}
}

func (r *EmailNotificationRelayRunner) Name() string {
	return EmailNotificationRelay
}
func (r *EmailNotificationRelayRunner) Description() string {
	return "Start the email notification relay process"
}
func (r *EmailNotificationRelayRunner) SetupFlags(_ *flag.FlagSet) {}
func (r *EmailNotificationRelayRunner) Run() error {
	return r.cmd.emailNotificationRelay()
}

// emailNotificationRelay starts the email notification relay process that polls
// the email_notification_outbox table and sends emails via SMTP.
func (cmd *Command) emailNotificationRelay() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctx = telemetry.InjectTracing(ctx, cmd.tracer)

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	done := make(chan struct{})

	dbPool, err := db.Open(context.Background(), cmd.db)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer dbPool.Close()

	store := db.NewStore(dbPool)

	emailClient := email.New(cmd.emailConfig)

	emailOutboxSvc := email_notification_outbox.NewService(
		cmd.log, store, uuidgen.NewV7(), clock.NewClock(),
	)

	emailRelay := relay.NewEmailRelay(
		cmd.log,
		emailOutboxSvc,
		emailClient,
		3*time.Second,
		emailBatchSize,
		emailMaxRetries,
	)

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
