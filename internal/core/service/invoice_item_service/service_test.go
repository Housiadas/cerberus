package invoice_item_service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Housiadas/cerberus/internal/core/domain/invoice_item"
	"github.com/Housiadas/cerberus/internal/core/service/invoice_item_service"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
)

func TestService_Create_Successful(t *testing.T) {
	ctx := context.Background()

	nii := invoice_item.NewInvoiceItem{
		InvoiceID:      uuid.New(),
		Description:    "Monthly subscription",
		Quantity:       1,
		UnitPriceCents: 9900,
		TaxRateID:      nil,
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := invoice_item.NewMockStorer(t)
	mStorer.EXPECT().Create(ctx, mock.AnythingOfType("invoice_item.InvoiceItem")).Return(nil)

	mUUID := uuidgen.NewMockGenerator(t)
	mUUID.EXPECT().Generate().Return(uuid.New(), nil)

	sut := invoice_item_service.New(mLogger, mStorer, mUUID)
	item, err := sut.Create(ctx, nii)

	assert.NoError(t, err)
	assert.Equal(t, nii.InvoiceID, item.InvoiceID())
	assert.Equal(t, nii.Description, item.Description())
	assert.NotZero(t, item.CreatedAt())
}

func TestService_Create_StorerError(t *testing.T) {
	ctx := context.Background()

	nii := invoice_item.NewInvoiceItem{
		InvoiceID:      uuid.New(),
		Description:    "Monthly subscription",
		Quantity:       1,
		UnitPriceCents: 9900,
		TaxRateID:      nil,
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := invoice_item.NewMockStorer(t)
	mStorer.EXPECT().Create(ctx, mock.AnythingOfType("invoice_item.InvoiceItem")).Return(errors.New("create error"))

	mUUID := uuidgen.NewMockGenerator(t)
	mUUID.EXPECT().Generate().Return(uuid.New(), nil)

	sut := invoice_item_service.New(mLogger, mStorer, mUUID)
	_, err := sut.Create(ctx, nii)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "create error")
}

func TestService_QueryByID_Successful(t *testing.T) {
	ctx := context.Background()
	itemID := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	expectedItem := invoice_item.New(
		itemID,
		uuid.New(),
		"Monthly subscription",
		1,
		9900,
		nil,
		9900,
		mTime,
	)

	mLogger := logger.NewMockLogger(t)

	mStorer := invoice_item.NewMockStorer(t)
	mStorer.EXPECT().QueryByID(ctx, itemID).Return(expectedItem, nil)

	sut := invoice_item_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	item, err := sut.QueryByID(ctx, itemID)

	assert.NoError(t, err)
	assert.Equal(t, expectedItem.ID(), item.ID())
	assert.Equal(t, expectedItem.Description(), item.Description())
}

func TestService_QueryByID_NotFound(t *testing.T) {
	ctx := context.Background()
	itemID := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")

	mLogger := logger.NewMockLogger(t)

	mStorer := invoice_item.NewMockStorer(t)
	mStorer.EXPECT().QueryByID(ctx, itemID).Return(invoice_item.InvoiceItem{}, invoice_item.ErrNotFound)

	sut := invoice_item_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	_, err := sut.QueryByID(ctx, itemID)

	assert.Error(t, err)
	assert.ErrorIs(t, err, invoice_item.ErrNotFound)
}

func TestService_Query_Successful(t *testing.T) {
	ctx := context.Background()
	filter := invoice_item.QueryFilter{}
	orderBy := order.By{Field: "created_at", Direction: "asc"}
	cur := cursor.Cursor{}

	expectedItems := []invoice_item.InvoiceItem{
		invoice_item.New(
			uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
			uuid.New(),
			"Monthly subscription",
			1,
			9900,
			nil,
			9900,
			time.Time{},
		),
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := invoice_item.NewMockStorer(t)
	mStorer.EXPECT().Query(ctx, filter, orderBy, cur).Return(expectedItems, nil)

	sut := invoice_item_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	items, err := sut.Query(ctx, filter, orderBy, cur)

	assert.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, expectedItems[0].Description(), items[0].Description())
}

func TestService_Query_Error(t *testing.T) {
	ctx := context.Background()
	filter := invoice_item.QueryFilter{}
	orderBy := order.By{Field: "created_at", Direction: "asc"}
	cur := cursor.Cursor{}

	mLogger := logger.NewMockLogger(t)

	mStorer := invoice_item.NewMockStorer(t)
	mStorer.EXPECT().Query(ctx, filter, orderBy, cur).Return(nil, errors.New("query error"))

	sut := invoice_item_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	_, err := sut.Query(ctx, filter, orderBy, cur)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "query error")
}
