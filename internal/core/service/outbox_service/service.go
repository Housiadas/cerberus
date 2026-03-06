// Package outbox_service is the service of the outbox domain.
package outbox_service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/outbox"
	"github.com/Housiadas/cerberus/pkg/clock"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
	"github.com/google/uuid"
)

// Service manages the set of APIs for outbox access.
type Service struct {
	log     logger.Logger
	storer  outbox.Storer
	uuidGen uuidgen.Generator
	clock   clock.Clock
}

// New constructs the service.
func New(
	log logger.Logger,
	storer outbox.Storer,
	uuidGen uuidgen.Generator,
	clock clock.Clock,
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
		return nil, fmt.Errorf("outbox transaction issue: %w", err)
	}

	svc := Service{
		log:     s.log,
		storer:  storer,
		uuidGen: s.uuidGen,
		clock:   s.clock,
	}

	return &svc, nil
}

// Create adds a new outbox entry.
func (s *Service) Create(ctx context.Context, no outbox.NewOutbox) error {
	id, err := s.uuidGen.Generate()
	if err != nil {
		return fmt.Errorf("uuid generate error: %w", err)
	}

	payload, err := json.Marshal(no.Payload)
	if err != nil {
		return fmt.Errorf("marshal payload error: %w", err)
	}

	now := s.clock.Now()
	o := outbox.New(id, no.EventType, no.AggregateID, no.Topic, payload, 0, now, nil)

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
) ([]outbox.Outbox, error) {
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
