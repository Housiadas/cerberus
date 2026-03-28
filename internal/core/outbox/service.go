// Package outbox_service is the service of the outbox domain.
package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type generator interface {
	Generate() (uuid.UUID, error)
}

type clock interface {
	Now() time.Time
}

type logger interface {
	Info(ctx context.Context, msg string, args ...any)
	Error(ctx context.Context, msg string, args ...any)
}

// Service manages the set of APIs for outbox access.
type Service struct {
	log     logger
	storer  Storer
	uuidGen generator
	clock   clock
}

// NewService constructs the service.
func NewService(
	log logger,
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

// Create adds a new outbox entry.
func (s *Service) Create(ctx context.Context, no NewOutbox) error {
	id, err := s.uuidGen.Generate()
	if err != nil {
		return fmt.Errorf("uuid generate error: %w", err)
	}

	payload, err := json.Marshal(no.Payload)
	if err != nil {
		return fmt.Errorf("marshal payload error: %w", err)
	}

	now := s.clock.Now()
	o := New(id, no.EventType, no.AggregateID, no.Topic, payload, 0, now, nil)

	err = s.storer.Create(ctx, o)
	if err != nil {
		return fmt.Errorf("create: %w", err)
	}

	return nil
}

// QueryUnprocessed returns unprocessed outbox entries up to the given limit,
// excluding entries that have reached the maximum retry count.
func (s *Service) QueryUnprocessed(
	ctx context.Context,
	limit int,
	maxRetries int,
) ([]Outbox, error) {
	entries, err := s.storer.QueryUnprocessed(ctx, limit, maxRetries)
	if err != nil {
		return nil, fmt.Errorf("query unprocessed: %w", err)
	}

	return entries, nil
}

// IncrementRetryCount increments the retry count for the given outbox entry IDs.
func (s *Service) IncrementRetryCount(ctx context.Context, ids []uuid.UUID) error {
	err := s.storer.IncrementRetryCount(ctx, ids)
	if err != nil {
		return fmt.Errorf("increment retry count: %w", err)
	}

	return nil
}

// MarkProcessed marks the given outbox entries as processed.
func (s *Service) MarkProcessed(ctx context.Context, ids []uuid.UUID) error {
	now := s.clock.Now()

	err := s.storer.MarkProcessed(ctx, ids, now)
	if err != nil {
		return fmt.Errorf("mark processed: %w", err)
	}

	return nil
}
