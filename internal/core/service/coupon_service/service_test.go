package coupon_service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Housiadas/cerberus/internal/core/domain/coupon"
	"github.com/Housiadas/cerberus/internal/core/service/coupon_service"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
)

func TestService_Create_Successful(t *testing.T) {
	ctx := context.Background()

	nc := coupon.NewCoupon{
		Code:           "SAVE10",
		DiscountType:   coupon.MustParse("percent"),
		DiscountValue:  10,
		Currency:       nil,
		MaxRedemptions: nil,
		ExpiresAt:      nil,
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := coupon.NewMockStorer(t)
	mStorer.EXPECT().Create(ctx, mock.AnythingOfType("coupon.Coupon")).Return(nil)

	mUUID := uuidgen.NewMockGenerator(t)
	mUUID.EXPECT().Generate().Return(uuid.New(), nil)

	sut := coupon_service.New(mLogger, mStorer, mUUID)
	cpn, err := sut.Create(ctx, nc)

	assert.NoError(t, err)
	assert.Equal(t, nc.Code, cpn.Code())
	assert.Equal(t, nc.DiscountType, cpn.DiscountType())
	assert.NotZero(t, cpn.CreatedAt())
	assert.NotZero(t, cpn.UpdatedAt())
}

func TestService_Create_StorerError(t *testing.T) {
	ctx := context.Background()

	nc := coupon.NewCoupon{
		Code:           "SAVE10",
		DiscountType:   coupon.MustParse("percent"),
		DiscountValue:  10,
		Currency:       nil,
		MaxRedemptions: nil,
		ExpiresAt:      nil,
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := coupon.NewMockStorer(t)
	mStorer.EXPECT().Create(ctx, mock.AnythingOfType("coupon.Coupon")).Return(errors.New("create error"))

	mUUID := uuidgen.NewMockGenerator(t)
	mUUID.EXPECT().Generate().Return(uuid.New(), nil)

	sut := coupon_service.New(mLogger, mStorer, mUUID)
	_, err := sut.Create(ctx, nc)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "create error")
}

func TestService_Update_Successful(t *testing.T) {
	ctx := context.Background()
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	existingCoupon := coupon.New(
		uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		"SAVE10",
		coupon.MustParse("percent"),
		10,
		nil,
		nil,
		0,
		true,
		nil,
		mTime,
		mTime,
	)

	newCode := "SAVE20"
	uc := coupon.UpdateCoupon{
		Code: &newCode,
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := coupon.NewMockStorer(t)
	mStorer.EXPECT().Update(ctx, mock.AnythingOfType("coupon.Coupon")).Return(nil)

	sut := coupon_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	cpn, err := sut.Update(ctx, existingCoupon, uc)

	assert.NoError(t, err)
	assert.Equal(t, newCode, cpn.Code())
	assert.True(t, cpn.UpdatedAt().After(mTime) || cpn.UpdatedAt().Equal(mTime))
}

func TestService_Update_StorerError(t *testing.T) {
	ctx := context.Background()
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	existingCoupon := coupon.New(
		uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		"SAVE10",
		coupon.MustParse("percent"),
		10,
		nil,
		nil,
		0,
		true,
		nil,
		mTime,
		mTime,
	)

	newCode := "SAVE20"
	uc := coupon.UpdateCoupon{
		Code: &newCode,
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := coupon.NewMockStorer(t)
	mStorer.EXPECT().Update(ctx, mock.AnythingOfType("coupon.Coupon")).Return(errors.New("update error"))

	sut := coupon_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	_, err := sut.Update(ctx, existingCoupon, uc)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update error")
}

func TestService_Delete_Successful(t *testing.T) {
	ctx := context.Background()

	cpn := coupon.New(
		uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		"SAVE10",
		coupon.MustParse("percent"),
		10,
		nil,
		nil,
		0,
		true,
		nil,
		time.Time{},
		time.Time{},
	)

	mLogger := logger.NewMockLogger(t)

	mStorer := coupon.NewMockStorer(t)
	mStorer.EXPECT().Delete(ctx, cpn).Return(nil)

	sut := coupon_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	err := sut.Delete(ctx, cpn)

	assert.NoError(t, err)
}

func TestService_Delete_Error(t *testing.T) {
	ctx := context.Background()

	cpn := coupon.New(
		uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		"SAVE10",
		coupon.MustParse("percent"),
		10,
		nil,
		nil,
		0,
		true,
		nil,
		time.Time{},
		time.Time{},
	)

	mLogger := logger.NewMockLogger(t)

	mStorer := coupon.NewMockStorer(t)
	mStorer.EXPECT().Delete(ctx, cpn).Return(errors.New("delete error"))

	sut := coupon_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	err := sut.Delete(ctx, cpn)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delete error")
}

func TestService_QueryByID_Successful(t *testing.T) {
	ctx := context.Background()
	couponID := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	expectedCoupon := coupon.New(
		couponID,
		"SAVE10",
		coupon.MustParse("percent"),
		10,
		nil,
		nil,
		0,
		true,
		nil,
		mTime,
		mTime,
	)

	mLogger := logger.NewMockLogger(t)

	mStorer := coupon.NewMockStorer(t)
	mStorer.EXPECT().QueryByID(ctx, couponID).Return(expectedCoupon, nil)

	sut := coupon_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	cpn, err := sut.QueryByID(ctx, couponID)

	assert.NoError(t, err)
	assert.Equal(t, expectedCoupon.ID(), cpn.ID())
	assert.Equal(t, expectedCoupon.Code(), cpn.Code())
}

func TestService_QueryByID_NotFound(t *testing.T) {
	ctx := context.Background()
	couponID := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")

	mLogger := logger.NewMockLogger(t)

	mStorer := coupon.NewMockStorer(t)
	mStorer.EXPECT().QueryByID(ctx, couponID).Return(coupon.Coupon{}, coupon.ErrNotFound)

	sut := coupon_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	_, err := sut.QueryByID(ctx, couponID)

	assert.Error(t, err)
	assert.ErrorIs(t, err, coupon.ErrNotFound)
}

func TestService_Query_Successful(t *testing.T) {
	ctx := context.Background()
	filter := coupon.QueryFilter{}
	orderBy := order.By{Field: "code", Direction: "asc"}
	cur := cursor.Cursor{}

	expectedCoupons := []coupon.Coupon{
		coupon.New(
			uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
			"SAVE10",
			coupon.MustParse("percent"),
			10,
			nil,
			nil,
			0,
			true,
			nil,
			time.Time{},
			time.Time{},
		),
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := coupon.NewMockStorer(t)
	mStorer.EXPECT().Query(ctx, filter, orderBy, cur).Return(expectedCoupons, nil)

	sut := coupon_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	coupons, err := sut.Query(ctx, filter, orderBy, cur)

	assert.NoError(t, err)
	assert.Len(t, coupons, 1)
	assert.Equal(t, expectedCoupons[0].Code(), coupons[0].Code())
}

func TestService_Query_Error(t *testing.T) {
	ctx := context.Background()
	filter := coupon.QueryFilter{}
	orderBy := order.By{Field: "code", Direction: "asc"}
	cur := cursor.Cursor{}

	mLogger := logger.NewMockLogger(t)

	mStorer := coupon.NewMockStorer(t)
	mStorer.EXPECT().Query(ctx, filter, orderBy, cur).Return(nil, errors.New("query error"))

	sut := coupon_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	_, err := sut.Query(ctx, filter, orderBy, cur)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "query error")
}
