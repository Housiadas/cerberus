package payment_service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Housiadas/cerberus/internal/core/domain/currency"
	"github.com/Housiadas/cerberus/internal/core/domain/payment"
	"github.com/Housiadas/cerberus/internal/core/service/payment_service"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
)

func TestService_Create_Successful(t *testing.T) {
	ctx := context.Background()

	np := payment.NewPayment{
		InvoiceID:       uuid.New(),
		PaymentMethodID: uuid.New(),
		AmountCents:     1000,
		Currency:        currency.MustParse("USD"),
		Status:          payment.MustParse("pending"),
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := payment.NewMockStorer(t)
	mStorer.EXPECT().Create(ctx, mock.AnythingOfType("payment.Payment")).Return(nil)

	mUUID := uuidgen.NewMockGenerator(t)
	mUUID.EXPECT().Generate().Return(uuid.New(), nil)

	sut := payment_service.New(mLogger, mStorer, mUUID)
	pmt, err := sut.Create(ctx, np)

	assert.NoError(t, err)
	assert.Equal(t, np.InvoiceID, pmt.InvoiceID())
	assert.Equal(t, np.Status, pmt.Status())
	assert.NotZero(t, pmt.CreatedAt())
	assert.NotZero(t, pmt.UpdatedAt())
}

func TestService_Create_StorerError(t *testing.T) {
	ctx := context.Background()

	np := payment.NewPayment{
		InvoiceID:       uuid.New(),
		PaymentMethodID: uuid.New(),
		AmountCents:     1000,
		Currency:        currency.MustParse("USD"),
		Status:          payment.MustParse("pending"),
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := payment.NewMockStorer(t)
	mStorer.EXPECT().Create(ctx, mock.AnythingOfType("payment.Payment")).Return(errors.New("create error"))

	mUUID := uuidgen.NewMockGenerator(t)
	mUUID.EXPECT().Generate().Return(uuid.New(), nil)

	sut := payment_service.New(mLogger, mStorer, mUUID)
	_, err := sut.Create(ctx, np)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "create error")
}

func TestService_Update_Successful(t *testing.T) {
	ctx := context.Background()
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	existingPayment := payment.New(
		uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		uuid.New(),
		uuid.New(),
		1000,
		currency.MustParse("USD"),
		payment.MustParse("pending"),
		nil,
		mTime,
		mTime,
	)

	newStatus := payment.MustParse("completed")
	up := payment.UpdatePayment{
		Status: &newStatus,
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := payment.NewMockStorer(t)
	mStorer.EXPECT().Update(ctx, mock.AnythingOfType("payment.Payment")).Return(nil)

	sut := payment_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	pmt, err := sut.Update(ctx, existingPayment, up)

	assert.NoError(t, err)
	assert.Equal(t, newStatus, pmt.Status())
	assert.True(t, pmt.UpdatedAt().After(mTime) || pmt.UpdatedAt().Equal(mTime))
}

func TestService_Update_StorerError(t *testing.T) {
	ctx := context.Background()
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	existingPayment := payment.New(
		uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		uuid.New(),
		uuid.New(),
		1000,
		currency.MustParse("USD"),
		payment.MustParse("pending"),
		nil,
		mTime,
		mTime,
	)

	newStatus := payment.MustParse("completed")
	up := payment.UpdatePayment{
		Status: &newStatus,
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := payment.NewMockStorer(t)
	mStorer.EXPECT().Update(ctx, mock.AnythingOfType("payment.Payment")).Return(errors.New("update error"))

	sut := payment_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	_, err := sut.Update(ctx, existingPayment, up)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update error")
}

func TestService_QueryByID_Successful(t *testing.T) {
	ctx := context.Background()
	paymentID := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	expectedPayment := payment.New(
		paymentID,
		uuid.New(),
		uuid.New(),
		1000,
		currency.MustParse("USD"),
		payment.MustParse("pending"),
		nil,
		mTime,
		mTime,
	)

	mLogger := logger.NewMockLogger(t)

	mStorer := payment.NewMockStorer(t)
	mStorer.EXPECT().QueryByID(ctx, paymentID).Return(expectedPayment, nil)

	sut := payment_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	pmt, err := sut.QueryByID(ctx, paymentID)

	assert.NoError(t, err)
	assert.Equal(t, expectedPayment.ID(), pmt.ID())
	assert.Equal(t, expectedPayment.Status(), pmt.Status())
}

func TestService_QueryByID_NotFound(t *testing.T) {
	ctx := context.Background()
	paymentID := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")

	mLogger := logger.NewMockLogger(t)

	mStorer := payment.NewMockStorer(t)
	mStorer.EXPECT().QueryByID(ctx, paymentID).Return(payment.Payment{}, payment.ErrNotFound)

	sut := payment_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	_, err := sut.QueryByID(ctx, paymentID)

	assert.Error(t, err)
	assert.ErrorIs(t, err, payment.ErrNotFound)
}

func TestService_Query_Successful(t *testing.T) {
	ctx := context.Background()
	filter := payment.QueryFilter{}
	orderBy := order.By{Field: "created_at", Direction: "asc"}
	cur := cursor.Cursor{}

	expectedPayments := []payment.Payment{
		payment.New(
			uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
			uuid.New(),
			uuid.New(),
			1000,
			currency.MustParse("USD"),
			payment.MustParse("pending"),
			nil,
			time.Time{},
			time.Time{},
		),
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := payment.NewMockStorer(t)
	mStorer.EXPECT().Query(ctx, filter, orderBy, cur).Return(expectedPayments, nil)

	sut := payment_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	payments, err := sut.Query(ctx, filter, orderBy, cur)

	assert.NoError(t, err)
	assert.Len(t, payments, 1)
	assert.Equal(t, expectedPayments[0].Status(), payments[0].Status())
}

func TestService_Query_Error(t *testing.T) {
	ctx := context.Background()
	filter := payment.QueryFilter{}
	orderBy := order.By{Field: "created_at", Direction: "asc"}
	cur := cursor.Cursor{}

	mLogger := logger.NewMockLogger(t)

	mStorer := payment.NewMockStorer(t)
	mStorer.EXPECT().Query(ctx, filter, orderBy, cur).Return(nil, errors.New("query error"))

	sut := payment_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	_, err := sut.Query(ctx, filter, orderBy, cur)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "query error")
}
