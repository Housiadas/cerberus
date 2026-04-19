package outbox

import (
	"context"
	"encoding/json"
	"fmt"

	db "github.com/Housiadas/cerberus/db/sqlc"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/mapper"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// Service manages the set of APIs for outbox access.
type Service struct {
	log     logger.Logger
	storer  storer
	uuidGen generator
	clock   clock
}

// NewService constructs the service.
func NewService(
	log logger.Logger,
	storer storer,
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
	params := toCreateOutboxParams(id, no, payload, now)

	_, err = s.storer.CreateOutbox(ctx, params)
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
	params := db.GetUnprocessedOutboxParams{
		Limit:      int32(limit),
		MaxRetries: int32(maxRetries),
	}

	dbEntries, err := s.storer.GetUnprocessedOutbox(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("query unprocessed: %w", err)
	}

	return mapper.MapSlice(dbEntries, toDomainOutbox), nil
}

// IncrementRetryCount increments the retry count for the given outbox entry IDs.
func (s *Service) IncrementRetryCount(ctx context.Context, ids []uuid.UUID) error {
	for _, id := range ids {
		err := s.storer.IncrementRetryOutbox(ctx, id)
		if err != nil {
			return fmt.Errorf("increment retry count: %w", err)
		}
	}

	return nil
}

// MarkProcessed marks the given outbox entries as processed.
func (s *Service) MarkProcessed(ctx context.Context, ids []uuid.UUID) error {
	now := s.clock.Now()

	params := db.MarkProcessedOutboxParams{
		ProcessedAt: pgtype.Timestamp{Time: now, Valid: true},
		Ids:         ids,
	}

	err := s.storer.MarkProcessedOutbox(ctx, params)
	if err != nil {
		return fmt.Errorf("mark processed: %w", err)
	}

	return nil
}
