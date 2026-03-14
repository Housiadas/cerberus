// Package usage_record_service provides internal access to usage record core.
package usage_record_service

import (
	"context"
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/usage_record"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
	"github.com/google/uuid"
)

// Service manages the set of APIs for usage record access.
type Service struct {
	log     logger.Logger
	storer  usage_record.Storer
	uuidGen uuidgen.Generator
}

// New constructor.
func New(log logger.Logger, storer usage_record.Storer, uuidGen uuidgen.Generator) *Service {
	return &Service{
		log:     log,
		storer:  storer,
		uuidGen: uuidGen,
	}
}

// NewWithTx constructs a new internal value that will use the
// specified transaction in any store-related calls.
func (c *Service) NewWithTx(tx pgsql.CommitRollbacker) (*Service, error) {
	storer, err := c.storer.NewWithTx(tx)
	if err != nil {
		return nil, fmt.Errorf("usage record transaction issue: %w", err)
	}

	bus := Service{
		log:     c.log,
		storer:  storer,
		uuidGen: c.uuidGen,
	}

	return &bus, nil
}

// Create adds a new usage_record.UsageRecord to the system.
func (c *Service) Create(
	ctx context.Context,
	nur usage_record.NewUsageRecord,
) (usage_record.UsageRecord, error) {
	id, err := c.uuidGen.Generate()
	if err != nil {
		return usage_record.UsageRecord{}, fmt.Errorf("usage record uuid generate: %w", err)
	}

	now := time.Now()
	ur := usage_record.New(
		id,
		nur.SubscriptionID,
		nur.Description,
		nur.Quantity,
		nur.RecordedAt,
		now,
	)

	err = c.storer.Create(ctx, ur)
	if err != nil {
		return usage_record.UsageRecord{}, fmt.Errorf("usage record create: %w", err)
	}

	return ur, nil
}

// QueryByID finds the usage record by the specified ID.
func (c *Service) QueryByID(
	ctx context.Context,
	usageRecordID uuid.UUID,
) (usage_record.UsageRecord, error) {
	ur, err := c.storer.QueryByID(ctx, usageRecordID)
	if err != nil {
		return usage_record.UsageRecord{}, fmt.Errorf(
			"query: usageRecordID[%s]: %w",
			usageRecordID,
			err,
		)
	}

	return ur, nil
}

// Query retrieves a list of existing usage records.
func (c *Service) Query(
	ctx context.Context,
	filter usage_record.QueryFilter,
	orderBy order.By,
	cur cursor.Cursor,
) ([]usage_record.UsageRecord, error) {
	urs, err := c.storer.Query(ctx, filter, orderBy, cur)
	if err != nil {
		return nil, fmt.Errorf("usage record query: %w", err)
	}

	return urs, nil
}
