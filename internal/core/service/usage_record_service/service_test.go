package usage_record_service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Housiadas/cerberus/internal/core/domain/usage_record"
	"github.com/Housiadas/cerberus/internal/core/service/usage_record_service"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
)

func TestService_Create_Successful(t *testing.T) {
	ctx := context.Background()

	nur := usage_record.NewUsageRecord{
		SubscriptionID: uuid.New(),
		Description:    "API calls",
		Quantity:       100,
		RecordedAt:     time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC),
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := usage_record.NewMockStorer(t)
	mStorer.EXPECT().Create(ctx, mock.AnythingOfType("usage_record.UsageRecord")).Return(nil)

	mUUID := uuidgen.NewMockGenerator(t)
	mUUID.EXPECT().Generate().Return(uuid.New(), nil)

	sut := usage_record_service.New(mLogger, mStorer, mUUID)
	ur, err := sut.Create(ctx, nur)

	assert.NoError(t, err)
	assert.Equal(t, nur.SubscriptionID, ur.SubscriptionID())
	assert.Equal(t, nur.Description, ur.Description())
	assert.NotZero(t, ur.CreatedAt())
}

func TestService_Create_StorerError(t *testing.T) {
	ctx := context.Background()

	nur := usage_record.NewUsageRecord{
		SubscriptionID: uuid.New(),
		Description:    "API calls",
		Quantity:       100,
		RecordedAt:     time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC),
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := usage_record.NewMockStorer(t)
	mStorer.EXPECT().Create(ctx, mock.AnythingOfType("usage_record.UsageRecord")).Return(errors.New("create error"))

	mUUID := uuidgen.NewMockGenerator(t)
	mUUID.EXPECT().Generate().Return(uuid.New(), nil)

	sut := usage_record_service.New(mLogger, mStorer, mUUID)
	_, err := sut.Create(ctx, nur)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "create error")
}

func TestService_QueryByID_Successful(t *testing.T) {
	ctx := context.Background()
	urID := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	expectedUR := usage_record.New(
		urID,
		uuid.New(),
		"API calls",
		100,
		mTime,
		mTime,
	)

	mLogger := logger.NewMockLogger(t)

	mStorer := usage_record.NewMockStorer(t)
	mStorer.EXPECT().QueryByID(ctx, urID).Return(expectedUR, nil)

	sut := usage_record_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	ur, err := sut.QueryByID(ctx, urID)

	assert.NoError(t, err)
	assert.Equal(t, expectedUR.ID(), ur.ID())
	assert.Equal(t, expectedUR.Description(), ur.Description())
}

func TestService_QueryByID_NotFound(t *testing.T) {
	ctx := context.Background()
	urID := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")

	mLogger := logger.NewMockLogger(t)

	mStorer := usage_record.NewMockStorer(t)
	mStorer.EXPECT().QueryByID(ctx, urID).Return(usage_record.UsageRecord{}, usage_record.ErrNotFound)

	sut := usage_record_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	_, err := sut.QueryByID(ctx, urID)

	assert.Error(t, err)
	assert.ErrorIs(t, err, usage_record.ErrNotFound)
}

func TestService_Query_Successful(t *testing.T) {
	ctx := context.Background()
	filter := usage_record.QueryFilter{}
	orderBy := order.By{Field: "created_at", Direction: "asc"}
	cur := cursor.Cursor{}

	expectedURs := []usage_record.UsageRecord{
		usage_record.New(
			uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
			uuid.New(),
			"API calls",
			100,
			time.Time{},
			time.Time{},
		),
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := usage_record.NewMockStorer(t)
	mStorer.EXPECT().Query(ctx, filter, orderBy, cur).Return(expectedURs, nil)

	sut := usage_record_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	urs, err := sut.Query(ctx, filter, orderBy, cur)

	assert.NoError(t, err)
	assert.Len(t, urs, 1)
	assert.Equal(t, expectedURs[0].Description(), urs[0].Description())
}

func TestService_Query_Error(t *testing.T) {
	ctx := context.Background()
	filter := usage_record.QueryFilter{}
	orderBy := order.By{Field: "created_at", Direction: "asc"}
	cur := cursor.Cursor{}

	mLogger := logger.NewMockLogger(t)

	mStorer := usage_record.NewMockStorer(t)
	mStorer.EXPECT().Query(ctx, filter, orderBy, cur).Return(nil, errors.New("query error"))

	sut := usage_record_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	_, err := sut.Query(ctx, filter, orderBy, cur)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "query error")
}
