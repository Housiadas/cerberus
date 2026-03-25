package command

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/outbox/outbox_repo"
	"github.com/Housiadas/cerberus/internal/core/service/outbox_service"
	"github.com/Housiadas/cerberus/internal/relay"
	"github.com/Housiadas/cerberus/pkg/clock"
	"github.com/Housiadas/cerberus/pkg/kafka"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/Housiadas/cerberus/pkg/telemetry"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
)

const (
	batchSize  = 100
	maxRetries = 3
)

// OutboxRelay starts the outbox relay process that polls the outbox table
// and publishes events to Kafka.
func (cmd *Command) OutboxRelay() error {
	db, err := pgsql.Open(cmd.db)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer db.Close()

	kafkaProducer, err := kafka.NewProducer(kafka.ProducerConfig{
		Brokers:          cmd.kafka.Brokers,
		AddressFamily:    cmd.kafka.AddressFamily,
		SecurityProtocol: cmd.kafka.SecurityProtocol,
		LogLevel:         cmd.kafka.LogLevel,
		MaxMessageBytes:  cmd.kafka.MaxMessageBytes,
	})
	if err != nil {
		return fmt.Errorf("creating Kafka producer: %w", err)
	}
	defer kafkaProducer.Close()

	outboxRepo := outbox_repo.NewStore(cmd.log, db)
	outboxSvc := outbox_service.New(cmd.log, outboxRepo, uuidgen.NewV7(), clock.NewClock())

	outboxRelay := relay.New(
		cmd.log,
		outboxSvc,
		kafkaProducer,
		3*time.Second,
		batchSize,
		maxRetries,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// inject tracing to ctx
	ctx = telemetry.InjectTracing(ctx, cmd.tracer)

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	done := make(chan struct{})

	go func() {
		outboxRelay.Start(ctx)
		close(done)
	}()

	cmd.log.Info(ctx, "startup", "status", "outbox relay started")

	sig := <-shutdown
	cmd.log.Info(ctx, "shutdown", "status", "outbox relay stopping", "signal", sig)

	cancel()
	<-done

	cmd.log.Info(ctx, "shutdown", "status", "outbox relay stopped")

	return nil
}
