package subscription_discount_service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Housiadas/cerberus/internal/core/domain/subscription_discount"
	"github.com/Housiadas/cerberus/internal/core/service/subscription_discount_service"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
)

func TestService_Create_Successful(t *testing.T) {
	ctx := context.Background()

	nsd := subscription_discount.NewSubscriptionDiscount{
		SubscriptionID: uuid.New(),
		CouponID:       uuid.New(),
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := subscription_discount.NewMockStorer(t)
	mStorer.EXPECT().Create(ctx, mock.AnythingOfType("subscription_discount.SubscriptionDiscount")).Return(nil)

	mUUID := uuidgen.NewMockGenerator(t)
	mUUID.EXPECT().Generate().Return(uuid.New(), nil)

	sut := subscription_discount_service.New(mLogger, mStorer, mUUID)
	sd, err := sut.Create(ctx, nsd)

	assert.NoError(t, err)
	assert.Equal(t, nsd.SubscriptionID, sd.SubscriptionID())
	assert.Equal(t, nsd.CouponID, sd.CouponID())
	assert.NotZero(t, sd.AppliedAt())
	assert.NotZero(t, sd.CreatedAt())
}

func TestService_Create_StorerError(t *testing.T) {
	ctx := context.Background()

	nsd := subscription_discount.NewSubscriptionDiscount{
		SubscriptionID: uuid.New(),
		CouponID:       uuid.New(),
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := subscription_discount.NewMockStorer(t)
	mStorer.EXPECT().Create(ctx, mock.AnythingOfType("subscription_discount.SubscriptionDiscount")).Return(errors.New("create error"))

	mUUID := uuidgen.NewMockGenerator(t)
	mUUID.EXPECT().Generate().Return(uuid.New(), nil)

	sut := subscription_discount_service.New(mLogger, mStorer, mUUID)
	_, err := sut.Create(ctx, nsd)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "create error")
}

func TestService_QueryByID_Successful(t *testing.T) {
	ctx := context.Background()
	sdID := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	expectedSD := subscription_discount.New(
		sdID,
		uuid.New(),
		uuid.New(),
		mTime,
		mTime,
	)

	mLogger := logger.NewMockLogger(t)

	mStorer := subscription_discount.NewMockStorer(t)
	mStorer.EXPECT().QueryByID(ctx, sdID).Return(expectedSD, nil)

	sut := subscription_discount_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	sd, err := sut.QueryByID(ctx, sdID)

	assert.NoError(t, err)
	assert.Equal(t, expectedSD.ID(), sd.ID())
	assert.Equal(t, expectedSD.SubscriptionID(), sd.SubscriptionID())
}

func TestService_QueryByID_NotFound(t *testing.T) {
	ctx := context.Background()
	sdID := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")

	mLogger := logger.NewMockLogger(t)

	mStorer := subscription_discount.NewMockStorer(t)
	mStorer.EXPECT().QueryByID(ctx, sdID).Return(subscription_discount.SubscriptionDiscount{}, subscription_discount.ErrNotFound)

	sut := subscription_discount_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	_, err := sut.QueryByID(ctx, sdID)

	assert.Error(t, err)
	assert.ErrorIs(t, err, subscription_discount.ErrNotFound)
}

func TestService_Query_Successful(t *testing.T) {
	ctx := context.Background()
	filter := subscription_discount.QueryFilter{}
	orderBy := order.By{Field: "created_at", Direction: "asc"}
	cur := cursor.Cursor{}

	expectedSDs := []subscription_discount.SubscriptionDiscount{
		subscription_discount.New(
			uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
			uuid.New(),
			uuid.New(),
			time.Time{},
			time.Time{},
		),
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := subscription_discount.NewMockStorer(t)
	mStorer.EXPECT().Query(ctx, filter, orderBy, cur).Return(expectedSDs, nil)

	sut := subscription_discount_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	sds, err := sut.Query(ctx, filter, orderBy, cur)

	assert.NoError(t, err)
	assert.Len(t, sds, 1)
	assert.Equal(t, expectedSDs[0].SubscriptionID(), sds[0].SubscriptionID())
}

func TestService_Query_Error(t *testing.T) {
	ctx := context.Background()
	filter := subscription_discount.QueryFilter{}
	orderBy := order.By{Field: "created_at", Direction: "asc"}
	cur := cursor.Cursor{}

	mLogger := logger.NewMockLogger(t)

	mStorer := subscription_discount.NewMockStorer(t)
	mStorer.EXPECT().Query(ctx, filter, orderBy, cur).Return(nil, errors.New("query error"))

	sut := subscription_discount_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	_, err := sut.Query(ctx, filter, orderBy, cur)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "query error")
}
