package usage_record_service

import (
	"context"
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/usage_record"
	"github.com/google/uuid"
)

// TestNewUsageRecords is a helper method for testing.
func TestNewUsageRecords(n int) []usage_record.NewUsageRecord {
	newURs := make([]usage_record.NewUsageRecord, n)

	for i := range n {
		nur := usage_record.NewUsageRecord{
			SubscriptionID: uuid.New(),
			Description:    fmt.Sprintf("Usage%d", i),
			Quantity:       i + 1,
			RecordedAt:     time.Now(),
		}

		newURs[i] = nur
	}

	return newURs
}

// TestSeedUsageRecords is a helper method for testing.
func TestSeedUsageRecords(
	ctx context.Context,
	n int,
	service *Service,
) ([]usage_record.UsageRecord, error) {
	newURs := TestNewUsageRecords(n)

	urs := make([]usage_record.UsageRecord, len(newURs))

	for i, nur := range newURs {
		ur, err := service.Create(ctx, nur)
		if err != nil {
			return nil, fmt.Errorf("seeding usage record: idx: %d : %w", i, err)
		}

		urs[i] = ur
	}

	return urs, nil
}
