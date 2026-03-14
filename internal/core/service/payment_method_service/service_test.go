package payment_method_service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Housiadas/cerberus/internal/core/domain/payment_method"
	"github.com/Housiadas/cerberus/internal/core/service/payment_method_service"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
)

func TestService_Create_Successful(t *testing.T) {
	ctx := context.Background()

	npm := payment_method.NewPaymentMethod{
		AccountID: uuid.New(),
		Type:      payment_method.MustParse("card"),
		Details:   "{}",
		IsDefault: false,
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := payment_method.NewMockStorer(t)
	mStorer.EXPECT().Create(ctx, mock.AnythingOfType("payment_method.PaymentMethod")).Return(nil)

	mUUID := uuidgen.NewMockGenerator(t)
	mUUID.EXPECT().Generate().Return(uuid.New(), nil)

	sut := payment_method_service.New(mLogger, mStorer, mUUID)
	pm, err := sut.Create(ctx, npm)

	assert.NoError(t, err)
	assert.Equal(t, npm.AccountID, pm.AccountID())
	assert.Equal(t, npm.Type, pm.Type())
	assert.NotZero(t, pm.CreatedAt())
	assert.NotZero(t, pm.UpdatedAt())
}

func TestService_Create_StorerError(t *testing.T) {
	ctx := context.Background()

	npm := payment_method.NewPaymentMethod{
		AccountID: uuid.New(),
		Type:      payment_method.MustParse("card"),
		Details:   "{}",
		IsDefault: false,
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := payment_method.NewMockStorer(t)
	mStorer.EXPECT().Create(ctx, mock.AnythingOfType("payment_method.PaymentMethod")).Return(errors.New("create error"))

	mUUID := uuidgen.NewMockGenerator(t)
	mUUID.EXPECT().Generate().Return(uuid.New(), nil)

	sut := payment_method_service.New(mLogger, mStorer, mUUID)
	_, err := sut.Create(ctx, npm)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "create error")
}

func TestService_Delete_Successful(t *testing.T) {
	ctx := context.Background()

	pm := payment_method.New(
		uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		uuid.New(),
		payment_method.MustParse("card"),
		"{}",
		false,
		time.Time{},
		time.Time{},
	)

	mLogger := logger.NewMockLogger(t)

	mStorer := payment_method.NewMockStorer(t)
	mStorer.EXPECT().Delete(ctx, pm).Return(nil)

	sut := payment_method_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	err := sut.Delete(ctx, pm)

	assert.NoError(t, err)
}

func TestService_Delete_Error(t *testing.T) {
	ctx := context.Background()

	pm := payment_method.New(
		uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		uuid.New(),
		payment_method.MustParse("card"),
		"{}",
		false,
		time.Time{},
		time.Time{},
	)

	mLogger := logger.NewMockLogger(t)

	mStorer := payment_method.NewMockStorer(t)
	mStorer.EXPECT().Delete(ctx, pm).Return(errors.New("delete error"))

	sut := payment_method_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	err := sut.Delete(ctx, pm)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delete error")
}

func TestService_QueryByID_Successful(t *testing.T) {
	ctx := context.Background()
	pmID := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	expectedPM := payment_method.New(
		pmID,
		uuid.New(),
		payment_method.MustParse("card"),
		"{}",
		true,
		mTime,
		mTime,
	)

	mLogger := logger.NewMockLogger(t)

	mStorer := payment_method.NewMockStorer(t)
	mStorer.EXPECT().QueryByID(ctx, pmID).Return(expectedPM, nil)

	sut := payment_method_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	pm, err := sut.QueryByID(ctx, pmID)

	assert.NoError(t, err)
	assert.Equal(t, expectedPM.ID(), pm.ID())
	assert.Equal(t, expectedPM.Type(), pm.Type())
}

func TestService_QueryByID_NotFound(t *testing.T) {
	ctx := context.Background()
	pmID := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")

	mLogger := logger.NewMockLogger(t)

	mStorer := payment_method.NewMockStorer(t)
	mStorer.EXPECT().QueryByID(ctx, pmID).Return(payment_method.PaymentMethod{}, payment_method.ErrNotFound)

	sut := payment_method_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	_, err := sut.QueryByID(ctx, pmID)

	assert.Error(t, err)
	assert.ErrorIs(t, err, payment_method.ErrNotFound)
}

func TestService_Query_Successful(t *testing.T) {
	ctx := context.Background()
	filter := payment_method.QueryFilter{}
	orderBy := order.By{Field: "created_at", Direction: "asc"}
	cur := cursor.Cursor{}

	expectedPMs := []payment_method.PaymentMethod{
		payment_method.New(
			uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
			uuid.New(),
			payment_method.MustParse("card"),
			"{}",
			true,
			time.Time{},
			time.Time{},
		),
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := payment_method.NewMockStorer(t)
	mStorer.EXPECT().Query(ctx, filter, orderBy, cur).Return(expectedPMs, nil)

	sut := payment_method_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	pms, err := sut.Query(ctx, filter, orderBy, cur)

	assert.NoError(t, err)
	assert.Len(t, pms, 1)
	assert.Equal(t, expectedPMs[0].Type(), pms[0].Type())
}

func TestService_Query_Error(t *testing.T) {
	ctx := context.Background()
	filter := payment_method.QueryFilter{}
	orderBy := order.By{Field: "created_at", Direction: "asc"}
	cur := cursor.Cursor{}

	mLogger := logger.NewMockLogger(t)

	mStorer := payment_method.NewMockStorer(t)
	mStorer.EXPECT().Query(ctx, filter, orderBy, cur).Return(nil, errors.New("query error"))

	sut := payment_method_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	_, err := sut.Query(ctx, filter, orderBy, cur)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "query error")
}
