package email_notification_outbox

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

// Service manages the set of APIs for email notification outbox access.
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
	params := toCreateNotificationOutboxParams(id, no, payload, now)

	_, err = s.storer.CreateNotificationOutbox(ctx, params)
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
	params := db.GetUnprocessedNotificationOutboxParams{
		Limit:      int32(limit),
		MaxRetries: int32(maxRetries),
	}

	dbEntries, err := s.storer.GetUnprocessedNotificationOutbox(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("query unprocessed: %w", err)
	}

	return mapper.MapSlice(dbEntries, toDomainEmailNotificationOutbox), nil
}

// IncrementRetryCount increments the retry count for the given email outbox entry IDs.
func (s *Service) IncrementRetryCount(ctx context.Context, ids []uuid.UUID) error {
	err := s.storer.IncrementRetryNotificationOutbox(ctx, ids)
	if err != nil {
		return fmt.Errorf("increment retry count: %w", err)
	}

	return nil
}

// MarkProcessed marks the given email outbox entries as processed.
func (s *Service) MarkProcessed(ctx context.Context, ids []uuid.UUID) error {
	now := s.clock.Now()

	params := db.MarkProcessedNotificationOutboxParams{
		ProcessedAt: pgtype.Timestamp{Time: now, Valid: true},
		Ids:         ids,
	}

	err := s.storer.MarkProcessedNotificationOutbox(ctx, params)
	if err != nil {
		return fmt.Errorf("mark processed: %w", err)
	}

	return nil
}
