package relay

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Housiadas/cerberus/internal/core/email_notification_outbox"
	"github.com/Housiadas/cerberus/pkg/email"
	"github.com/Housiadas/cerberus/pkg/telemetry"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
)

// emailPayload is the JSON structure expected in the outbox payload field.
type emailPayload struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// EmailRelay polls in the email_notification_outbox table and sends emails.
type EmailRelay struct {
	log            logger
	emailOutboxSvc *email_notification_outbox.Service
	emailClient    *email.Client
	interval       time.Duration
	batchSize      int
	maxRetries     int
}

// NewEmailRelay constructs a new EmailRelay.
func NewEmailRelay(
	log logger,
	emailOutboxSvc *email_notification_outbox.Service,
	emailClient *email.Client,
	interval time.Duration,
	batchSize int,
	maxRetries int,
) *EmailRelay {
	return &EmailRelay{
		log:            log,
		emailOutboxSvc: emailOutboxSvc,
		emailClient:    emailClient,
		interval:       interval,
		batchSize:      batchSize,
		maxRetries:     maxRetries,
	}
}

// Start begins the relay polling loop. It blocks until the context is canceled.
func (r *EmailRelay) Start(ctx context.Context) {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	r.log.Info(
		ctx,
		"email_relay",
		"status", "started",
		"interval", r.interval.String(),
		"batchSize", r.batchSize,
		"maxRetries", r.maxRetries,
	)

	for {
		select {
		case <-ctx.Done():
			r.log.Info(ctx, "email_relay", "status", "stopped")

			return
		case <-ticker.C:
			r.processBatch(ctx)
		}
	}
}

func (r *EmailRelay) processBatch(ctx context.Context) {
	entries, err := r.emailOutboxSvc.QueryUnprocessed(ctx, r.batchSize, r.maxRetries)
	if err != nil {
		r.log.Error(ctx, "email_relay", "status", "query unprocessed error", "msg", err)

		return
	}

	if len(entries) == 0 {
		return
	}

	var (
		failedIDs    []uuid.UUID
		processedIDs []uuid.UUID
	)

	for _, entry := range entries {
		var payload emailPayload

		unmarshalErr := json.Unmarshal(entry.Payload(), &payload)
		if unmarshalErr != nil {
			r.log.Error(
				ctx,
				"email_relay",
				"status", "unmarshal payload error",
				"event_id", entry.ID(),
				"msg", unmarshalErr,
			)
			failedIDs = append(failedIDs, entry.ID())

			continue
		}

		sendErr := r.emailClient.Send(ctx, entry.ToEmail(), payload.Subject, payload.Body)
		if sendErr != nil {
			r.log.Error(
				ctx,
				"email_relay",
				"status", "send email error",
				"event_id", entry.ID(),
				"to", entry.ToEmail(),
				"msg", sendErr,
			)
			failedIDs = append(failedIDs, entry.ID())

			continue
		}

		processedIDs = append(processedIDs, entry.ID())
	}

	telemetry.AddSpan(ctx,
		"email_relay.processedIDs",
		attribute.Int("processedIDs", len(processedIDs)),
	)

	r.log.Info(
		ctx,
		"email_relay",
		"status", "batch processed",
		"total", len(entries),
		"sent", len(processedIDs),
		"failed", len(failedIDs),
	)

	if len(failedIDs) > 0 {
		telemetry.AddSpan(ctx, "email_relay.failedIDs", attribute.Int("failedIDs", len(failedIDs)))

		retryErr := r.emailOutboxSvc.IncrementRetryCount(ctx, failedIDs)
		if retryErr != nil {
			r.log.Error(
				ctx,
				"email_relay",
				"status", "increment retry count error",
				"msg", retryErr,
			)
		}
	}

	if len(processedIDs) == 0 {
		return
	}

	markErr := r.emailOutboxSvc.MarkProcessed(ctx, processedIDs)
	if markErr != nil {
		r.log.Error(ctx, "email_relay", "status", "mark processed error", "msg", markErr)
	}
}
