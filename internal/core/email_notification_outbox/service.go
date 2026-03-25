// Package email_notification_outbox_service provides access to the email notification outbox domain.
package email_notification_outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/pkg/pgsql"
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

// Service manages the set of APIs for email notification outbox access.
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

// NewWithTx constructs a new Service value that will use the
// specified transaction in any store-related calls.
func (s *Service) NewWithTx(tx pgsql.CommitRollbacker) (*Service, error) {
	storer, err := s.storer.NewWithTx(tx)
	if err != nil {
		return nil, fmt.Errorf("email_notification_outbox transaction issue: %w", err)
	}

	svc := Service{
		log:     s.log,
		storer:  storer,
		uuidGen: s.uuidGen,
		clock:   s.clock,
	}

	return &svc, nil
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
