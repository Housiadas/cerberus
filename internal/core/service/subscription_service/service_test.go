package subscription_service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Housiadas/cerberus/internal/core/domain/subscription"
	"github.com/Housiadas/cerberus/internal/core/service/subscription_service"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
)

func TestService_Create_Successful(t *testing.T) {
	ctx := context.Background()

	ns := subscription.NewSubscription{
		AccountID:          uuid.New(),
		PlanID:             uuid.New(),
		Status:             subscription.MustParse("active"),
		CurrentPeriodStart: time.Now(),
		CurrentPeriodEnd:   time.Now().Add(30 * 24 * time.Hour),
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := subscription.NewMockStorer(t)
	mStorer.EXPECT().Create(ctx, mock.AnythingOfType("subscription.Subscription")).Return(nil)

	mUUID := uuidgen.NewMockGenerator(t)
	mUUID.EXPECT().Generate().Return(uuid.New(), nil)

	sut := subscription_service.New(mLogger, mStorer, mUUID)
	sub, err := sut.Create(ctx, ns)

	assert.NoError(t, err)
	assert.Equal(t, ns.AccountID, sub.AccountID())
	assert.Equal(t, ns.PlanID, sub.PlanID())
	assert.NotZero(t, sub.CreatedAt())
	assert.NotZero(t, sub.UpdatedAt())
}

func TestService_Create_StorerError(t *testing.T) {
	ctx := context.Background()

	ns := subscription.NewSubscription{
		AccountID:          uuid.New(),
		PlanID:             uuid.New(),
		Status:             subscription.MustParse("active"),
		CurrentPeriodStart: time.Now(),
		CurrentPeriodEnd:   time.Now().Add(30 * 24 * time.Hour),
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := subscription.NewMockStorer(t)
	mStorer.EXPECT().Create(ctx, mock.AnythingOfType("subscription.Subscription")).Return(errors.New("create error"))

	mUUID := uuidgen.NewMockGenerator(t)
	mUUID.EXPECT().Generate().Return(uuid.New(), nil)

	sut := subscription_service.New(mLogger, mStorer, mUUID)
	_, err := sut.Create(ctx, ns)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "create error")
}

func TestService_Update_Successful(t *testing.T) {
	ctx := context.Background()
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	existingSub := subscription.New(
		uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		uuid.New(),
		uuid.New(),
		subscription.MustParse("active"),
		mTime,
		mTime.Add(30*24*time.Hour),
		mTime,
		mTime,
	)

	newStatus := subscription.MustParse("canceled")
	us := subscription.UpdateSubscription{
		Status: &newStatus,
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := subscription.NewMockStorer(t)
	mStorer.EXPECT().Update(ctx, mock.AnythingOfType("subscription.Subscription")).Return(nil)

	sut := subscription_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	sub, err := sut.Update(ctx, existingSub, us)

	assert.NoError(t, err)
	assert.Equal(t, newStatus, sub.Status())
	assert.True(t, sub.UpdatedAt().After(mTime) || sub.UpdatedAt().Equal(mTime))
}

func TestService_Update_StorerError(t *testing.T) {
	ctx := context.Background()
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	existingSub := subscription.New(
		uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		uuid.New(),
		uuid.New(),
		subscription.MustParse("active"),
		mTime,
		mTime.Add(30*24*time.Hour),
		mTime,
		mTime,
	)

	newStatus := subscription.MustParse("canceled")
	us := subscription.UpdateSubscription{
		Status: &newStatus,
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := subscription.NewMockStorer(t)
	mStorer.EXPECT().Update(ctx, mock.AnythingOfType("subscription.Subscription")).Return(errors.New("update error"))

	sut := subscription_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	_, err := sut.Update(ctx, existingSub, us)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update error")
}

func TestService_QueryByID_Successful(t *testing.T) {
	ctx := context.Background()
	subscriptionID := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	expectedSub := subscription.New(
		subscriptionID,
		uuid.New(),
		uuid.New(),
		subscription.MustParse("active"),
		mTime,
		mTime.Add(30*24*time.Hour),
		mTime,
		mTime,
	)

	mLogger := logger.NewMockLogger(t)

	mStorer := subscription.NewMockStorer(t)
	mStorer.EXPECT().QueryByID(ctx, subscriptionID).Return(expectedSub, nil)

	sut := subscription_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	sub, err := sut.QueryByID(ctx, subscriptionID)

	assert.NoError(t, err)
	assert.Equal(t, expectedSub.ID(), sub.ID())
	assert.Equal(t, expectedSub.Status(), sub.Status())
}

func TestService_QueryByID_NotFound(t *testing.T) {
	ctx := context.Background()
	subscriptionID := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")

	mLogger := logger.NewMockLogger(t)

	mStorer := subscription.NewMockStorer(t)
	mStorer.EXPECT().QueryByID(ctx, subscriptionID).Return(subscription.Subscription{}, subscription.ErrNotFound)

	sut := subscription_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	_, err := sut.QueryByID(ctx, subscriptionID)

	assert.Error(t, err)
	assert.ErrorIs(t, err, subscription.ErrNotFound)
}

func TestService_Query_Successful(t *testing.T) {
	ctx := context.Background()
	filter := subscription.QueryFilter{}
	orderBy := order.By{Field: "created_at", Direction: "asc"}
	cur := cursor.Cursor{}

	expectedSubs := []subscription.Subscription{
		subscription.New(
			uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
			uuid.New(),
			uuid.New(),
			subscription.MustParse("active"),
			time.Time{},
			time.Time{},
			time.Time{},
			time.Time{},
		),
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := subscription.NewMockStorer(t)
	mStorer.EXPECT().Query(ctx, filter, orderBy, cur).Return(expectedSubs, nil)

	sut := subscription_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	subs, err := sut.Query(ctx, filter, orderBy, cur)

	assert.NoError(t, err)
	assert.Len(t, subs, 1)
	assert.Equal(t, expectedSubs[0].Status(), subs[0].Status())
}

func TestService_Query_Error(t *testing.T) {
	ctx := context.Background()
	filter := subscription.QueryFilter{}
	orderBy := order.By{Field: "created_at", Direction: "asc"}
	cur := cursor.Cursor{}

	mLogger := logger.NewMockLogger(t)

	mStorer := subscription.NewMockStorer(t)
	mStorer.EXPECT().Query(ctx, filter, orderBy, cur).Return(nil, errors.New("query error"))

	sut := subscription_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	_, err := sut.Query(ctx, filter, orderBy, cur)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "query error")
}
