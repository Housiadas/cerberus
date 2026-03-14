package refund_service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Housiadas/cerberus/internal/core/domain/payment"
	"github.com/Housiadas/cerberus/internal/core/domain/refund"
	"github.com/Housiadas/cerberus/internal/core/service/refund_service"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
)

func TestService_Create_Successful(t *testing.T) {
	ctx := context.Background()

	nr := refund.NewRefund{
		PaymentID:   uuid.New(),
		AmountCents: 5000,
		Reason:      "Customer request",
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := refund.NewMockStorer(t)
	mStorer.EXPECT().Create(ctx, mock.AnythingOfType("refund.Refund")).Return(nil)

	mUUID := uuidgen.NewMockGenerator(t)
	mUUID.EXPECT().Generate().Return(uuid.New(), nil)

	sut := refund_service.New(mLogger, mStorer, mUUID)
	ref, err := sut.Create(ctx, nr)

	assert.NoError(t, err)
	assert.Equal(t, nr.PaymentID, ref.PaymentID())
	assert.Equal(t, nr.AmountCents, ref.AmountCents())
	assert.Equal(t, payment.MustParse("pending"), ref.Status())
	assert.NotZero(t, ref.CreatedAt())
	assert.NotZero(t, ref.UpdatedAt())
}

func TestService_Create_StorerError(t *testing.T) {
	ctx := context.Background()

	nr := refund.NewRefund{
		PaymentID:   uuid.New(),
		AmountCents: 5000,
		Reason:      "Customer request",
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := refund.NewMockStorer(t)
	mStorer.EXPECT().Create(ctx, mock.AnythingOfType("refund.Refund")).Return(errors.New("create error"))

	mUUID := uuidgen.NewMockGenerator(t)
	mUUID.EXPECT().Generate().Return(uuid.New(), nil)

	sut := refund_service.New(mLogger, mStorer, mUUID)
	_, err := sut.Create(ctx, nr)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "create error")
}

func TestService_QueryByID_Successful(t *testing.T) {
	ctx := context.Background()
	refID := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	expectedRef := refund.New(
		refID,
		uuid.New(),
		5000,
		"Customer request",
		payment.MustParse("pending"),
		mTime,
		mTime,
	)

	mLogger := logger.NewMockLogger(t)

	mStorer := refund.NewMockStorer(t)
	mStorer.EXPECT().QueryByID(ctx, refID).Return(expectedRef, nil)

	sut := refund_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	ref, err := sut.QueryByID(ctx, refID)

	assert.NoError(t, err)
	assert.Equal(t, expectedRef.ID(), ref.ID())
	assert.Equal(t, expectedRef.Reason(), ref.Reason())
}

func TestService_QueryByID_NotFound(t *testing.T) {
	ctx := context.Background()
	refID := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")

	mLogger := logger.NewMockLogger(t)

	mStorer := refund.NewMockStorer(t)
	mStorer.EXPECT().QueryByID(ctx, refID).Return(refund.Refund{}, refund.ErrNotFound)

	sut := refund_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	_, err := sut.QueryByID(ctx, refID)

	assert.Error(t, err)
	assert.ErrorIs(t, err, refund.ErrNotFound)
}

func TestService_Query_Successful(t *testing.T) {
	ctx := context.Background()
	filter := refund.QueryFilter{}
	orderBy := order.By{Field: "created_at", Direction: "asc"}
	cur := cursor.Cursor{}

	expectedRefs := []refund.Refund{
		refund.New(
			uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
			uuid.New(),
			5000,
			"Customer request",
			payment.MustParse("pending"),
			time.Time{},
			time.Time{},
		),
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := refund.NewMockStorer(t)
	mStorer.EXPECT().Query(ctx, filter, orderBy, cur).Return(expectedRefs, nil)

	sut := refund_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	refs, err := sut.Query(ctx, filter, orderBy, cur)

	assert.NoError(t, err)
	assert.Len(t, refs, 1)
	assert.Equal(t, expectedRefs[0].Reason(), refs[0].Reason())
}

func TestService_Query_Error(t *testing.T) {
	ctx := context.Background()
	filter := refund.QueryFilter{}
	orderBy := order.By{Field: "created_at", Direction: "asc"}
	cur := cursor.Cursor{}

	mLogger := logger.NewMockLogger(t)

	mStorer := refund.NewMockStorer(t)
	mStorer.EXPECT().Query(ctx, filter, orderBy, cur).Return(nil, errors.New("query error"))

	sut := refund_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	_, err := sut.Query(ctx, filter, orderBy, cur)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "query error")
}
