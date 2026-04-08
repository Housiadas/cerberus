package email_notification_outbox

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/google/uuid"
)

// Service manages the set of APIs for email notification outbox access.
type Service struct {
	log     logger.Logger
	storer  Storer
	uuidGen generator
	clock   clock
}

// NewService constructs the service.
func NewService(
	log logger.Logger,
	storer Storer,
	uuidGen generator,
	clock clock,
) *Service {
	return &Service{
		log:     log,
		storer:  storer,
		uuidGen: uuidGen,
		clock:   clock,
	}
}

// Create adds a new email notification outbox entry.
func (s *Service) Create(
	ctx context.Context,
	no NewEmailNotificationOutbox,
) error {
	id, err := s.uuidGen.Generate()
	if err != nil {
		return fmt.Errorf("uuid generate error: %w", err)
	}

	payload, err := json.Marshal(no.Payload)
	if err != nil {
		return fmt.Errorf("marshal payload error: %w", err)
	}

	now := s.clock.Now()
	e := New(id, no.EventType, no.ToEmail, payload, 0, now, nil)

	err = s.storer.Create(ctx, e)
	if err != nil {
		return fmt.Errorf("create: %w", err)
	}

	return nil
}

// QueryUnprocessed returns unprocessed email notification entries up to the given limit.
func (s *Service) QueryUnprocessed(
	ctx context.Context,
	limit int,
	maxRetries int,
) ([]EmailNotificationOutbox, error) {
	entries, err := s.storer.QueryUnprocessed(ctx, limit, maxRetries)
	if err != nil {
		return nil, fmt.Errorf("query unprocessed: %w", err)
	}

	return entries, nil
}

// IncrementRetryCount increments the retry count for the given email outbox entry IDs.
func (s *Service) IncrementRetryCount(ctx context.Context, ids []uuid.UUID) error {
	err := s.storer.IncrementRetryCount(ctx, ids)
	if err != nil {
		return fmt.Errorf("increment retry count: %w", err)
	}

	return nil
}

// MarkProcessed marks the given email outbox entries as processed.
func (s *Service) MarkProcessed(ctx context.Context, ids []uuid.UUID) error {
	now := s.clock.Now()

	err := s.storer.MarkProcessed(ctx, ids, now)
	if err != nil {
		return fmt.Errorf("mark processed: %w", err)
	}

	return nil
}
