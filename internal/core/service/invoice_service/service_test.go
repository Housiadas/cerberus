package invoice_service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Housiadas/cerberus/internal/core/domain/currency"
	"github.com/Housiadas/cerberus/internal/core/domain/invoice"
	"github.com/Housiadas/cerberus/internal/core/service/invoice_service"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
)

func TestService_Create_Successful(t *testing.T) {
	ctx := context.Background()

	ni := invoice.NewInvoice{
		AccountID:      uuid.New(),
		SubscriptionID: nil,
		Status:         invoice.MustParse("draft"),
		Currency:       currency.MustParse("USD"),
		DueDate:        nil,
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := invoice.NewMockStorer(t)
	mStorer.EXPECT().Create(ctx, mock.AnythingOfType("invoice.Invoice")).Return(nil)

	mUUID := uuidgen.NewMockGenerator(t)
	mUUID.EXPECT().Generate().Return(uuid.New(), nil)

	sut := invoice_service.New(mLogger, mStorer, mUUID)
	inv, err := sut.Create(ctx, ni)

	assert.NoError(t, err)
	assert.Equal(t, ni.AccountID, inv.AccountID())
	assert.Equal(t, ni.Status, inv.Status())
	assert.NotZero(t, inv.CreatedAt())
	assert.NotZero(t, inv.UpdatedAt())
}

func TestService_Create_StorerError(t *testing.T) {
	ctx := context.Background()

	ni := invoice.NewInvoice{
		AccountID:      uuid.New(),
		SubscriptionID: nil,
		Status:         invoice.MustParse("draft"),
		Currency:       currency.MustParse("USD"),
		DueDate:        nil,
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := invoice.NewMockStorer(t)
	mStorer.EXPECT().Create(ctx, mock.AnythingOfType("invoice.Invoice")).Return(errors.New("create error"))

	mUUID := uuidgen.NewMockGenerator(t)
	mUUID.EXPECT().Generate().Return(uuid.New(), nil)

	sut := invoice_service.New(mLogger, mStorer, mUUID)
	_, err := sut.Create(ctx, ni)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "create error")
}

func TestService_Update_Successful(t *testing.T) {
	ctx := context.Background()
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	existingInvoice := invoice.New(
		uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		uuid.New(),
		nil,
		invoice.MustParse("draft"),
		currency.MustParse("USD"),
		nil,
		nil,
		nil,
		mTime,
		mTime,
	)

	newStatus := invoice.MustParse("issued")
	ui := invoice.UpdateInvoice{
		Status: &newStatus,
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := invoice.NewMockStorer(t)
	mStorer.EXPECT().Update(ctx, mock.AnythingOfType("invoice.Invoice")).Return(nil)

	sut := invoice_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	inv, err := sut.Update(ctx, existingInvoice, ui)

	assert.NoError(t, err)
	assert.Equal(t, newStatus, inv.Status())
	assert.True(t, inv.UpdatedAt().After(mTime) || inv.UpdatedAt().Equal(mTime))
}

func TestService_Update_StorerError(t *testing.T) {
	ctx := context.Background()
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	existingInvoice := invoice.New(
		uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		uuid.New(),
		nil,
		invoice.MustParse("draft"),
		currency.MustParse("USD"),
		nil,
		nil,
		nil,
		mTime,
		mTime,
	)

	newStatus := invoice.MustParse("issued")
	ui := invoice.UpdateInvoice{
		Status: &newStatus,
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := invoice.NewMockStorer(t)
	mStorer.EXPECT().Update(ctx, mock.AnythingOfType("invoice.Invoice")).Return(errors.New("update error"))

	sut := invoice_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	_, err := sut.Update(ctx, existingInvoice, ui)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update error")
}

func TestService_QueryByID_Successful(t *testing.T) {
	ctx := context.Background()
	invoiceID := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	expectedInvoice := invoice.New(
		invoiceID,
		uuid.New(),
		nil,
		invoice.MustParse("draft"),
		currency.MustParse("USD"),
		nil,
		nil,
		nil,
		mTime,
		mTime,
	)

	mLogger := logger.NewMockLogger(t)

	mStorer := invoice.NewMockStorer(t)
	mStorer.EXPECT().QueryByID(ctx, invoiceID).Return(expectedInvoice, nil)

	sut := invoice_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	inv, err := sut.QueryByID(ctx, invoiceID)

	assert.NoError(t, err)
	assert.Equal(t, expectedInvoice.ID(), inv.ID())
	assert.Equal(t, expectedInvoice.Status(), inv.Status())
}

func TestService_QueryByID_NotFound(t *testing.T) {
	ctx := context.Background()
	invoiceID := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")

	mLogger := logger.NewMockLogger(t)

	mStorer := invoice.NewMockStorer(t)
	mStorer.EXPECT().QueryByID(ctx, invoiceID).Return(invoice.Invoice{}, invoice.ErrNotFound)

	sut := invoice_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	_, err := sut.QueryByID(ctx, invoiceID)

	assert.Error(t, err)
	assert.ErrorIs(t, err, invoice.ErrNotFound)
}

func TestService_Query_Successful(t *testing.T) {
	ctx := context.Background()
	filter := invoice.QueryFilter{}
	orderBy := order.By{Field: "created_at", Direction: "asc"}
	cur := cursor.Cursor{}

	expectedInvoices := []invoice.Invoice{
		invoice.New(
			uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
			uuid.New(),
			nil,
			invoice.MustParse("draft"),
			currency.MustParse("USD"),
			nil,
			nil,
			nil,
			time.Time{},
			time.Time{},
		),
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := invoice.NewMockStorer(t)
	mStorer.EXPECT().Query(ctx, filter, orderBy, cur).Return(expectedInvoices, nil)

	sut := invoice_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	invoices, err := sut.Query(ctx, filter, orderBy, cur)

	assert.NoError(t, err)
	assert.Len(t, invoices, 1)
	assert.Equal(t, expectedInvoices[0].Status(), invoices[0].Status())
}

func TestService_Query_Error(t *testing.T) {
	ctx := context.Background()
	filter := invoice.QueryFilter{}
	orderBy := order.By{Field: "created_at", Direction: "asc"}
	cur := cursor.Cursor{}

	mLogger := logger.NewMockLogger(t)

	mStorer := invoice.NewMockStorer(t)
	mStorer.EXPECT().Query(ctx, filter, orderBy, cur).Return(nil, errors.New("query error"))

	sut := invoice_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	_, err := sut.Query(ctx, filter, orderBy, cur)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "query error")
}
