package tax_rate_service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Housiadas/cerberus/internal/core/domain/country"
	"github.com/Housiadas/cerberus/internal/core/domain/tax_rate"
	"github.com/Housiadas/cerberus/internal/core/service/tax_rate_service"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
)

func TestService_Create_Successful(t *testing.T) {
	ctx := context.Background()

	nt := tax_rate.NewTaxRate{
		Name:        "VAT",
		Percentage:  20.0,
		Country:     country.MustParse("US"),
		Description: "Standard VAT",
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := tax_rate.NewMockStorer(t)
	mStorer.EXPECT().Create(ctx, mock.AnythingOfType("tax_rate.TaxRate")).Return(nil)

	mUUID := uuidgen.NewMockGenerator(t)
	mUUID.EXPECT().Generate().Return(uuid.New(), nil)

	sut := tax_rate_service.New(mLogger, mStorer, mUUID)
	tr, err := sut.Create(ctx, nt)

	assert.NoError(t, err)
	assert.Equal(t, nt.Name, tr.Name())
	assert.Equal(t, nt.Percentage, tr.Percentage())
	assert.NotZero(t, tr.CreatedAt())
	assert.NotZero(t, tr.UpdatedAt())
}

func TestService_Create_StorerError(t *testing.T) {
	ctx := context.Background()

	nt := tax_rate.NewTaxRate{
		Name:        "VAT",
		Percentage:  20.0,
		Country:     country.MustParse("US"),
		Description: "Standard VAT",
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := tax_rate.NewMockStorer(t)
	mStorer.EXPECT().Create(ctx, mock.AnythingOfType("tax_rate.TaxRate")).Return(errors.New("create error"))

	mUUID := uuidgen.NewMockGenerator(t)
	mUUID.EXPECT().Generate().Return(uuid.New(), nil)

	sut := tax_rate_service.New(mLogger, mStorer, mUUID)
	_, err := sut.Create(ctx, nt)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "create error")
}

func TestService_Update_Successful(t *testing.T) {
	ctx := context.Background()
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	existingTaxRate := tax_rate.New(
		uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		"VAT",
		20.0,
		country.MustParse("US"),
		"Standard VAT",
		true,
		mTime,
		mTime,
	)

	newName := "Reduced VAT"
	ut := tax_rate.UpdateTaxRate{
		Name: &newName,
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := tax_rate.NewMockStorer(t)
	mStorer.EXPECT().Update(ctx, mock.AnythingOfType("tax_rate.TaxRate")).Return(nil)

	sut := tax_rate_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	tr, err := sut.Update(ctx, existingTaxRate, ut)

	assert.NoError(t, err)
	assert.Equal(t, newName, tr.Name())
	assert.True(t, tr.UpdatedAt().After(mTime) || tr.UpdatedAt().Equal(mTime))
}

func TestService_Update_StorerError(t *testing.T) {
	ctx := context.Background()
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	existingTaxRate := tax_rate.New(
		uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		"VAT",
		20.0,
		country.MustParse("US"),
		"Standard VAT",
		true,
		mTime,
		mTime,
	)

	newName := "Reduced VAT"
	ut := tax_rate.UpdateTaxRate{
		Name: &newName,
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := tax_rate.NewMockStorer(t)
	mStorer.EXPECT().Update(ctx, mock.AnythingOfType("tax_rate.TaxRate")).Return(errors.New("update error"))

	sut := tax_rate_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	_, err := sut.Update(ctx, existingTaxRate, ut)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update error")
}

func TestService_Delete_Successful(t *testing.T) {
	ctx := context.Background()

	tr := tax_rate.New(
		uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		"VAT",
		20.0,
		country.MustParse("US"),
		"Standard VAT",
		true,
		time.Time{},
		time.Time{},
	)

	mLogger := logger.NewMockLogger(t)

	mStorer := tax_rate.NewMockStorer(t)
	mStorer.EXPECT().Delete(ctx, tr).Return(nil)

	sut := tax_rate_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	err := sut.Delete(ctx, tr)

	assert.NoError(t, err)
}

func TestService_Delete_Error(t *testing.T) {
	ctx := context.Background()

	tr := tax_rate.New(
		uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		"VAT",
		20.0,
		country.MustParse("US"),
		"Standard VAT",
		true,
		time.Time{},
		time.Time{},
	)

	mLogger := logger.NewMockLogger(t)

	mStorer := tax_rate.NewMockStorer(t)
	mStorer.EXPECT().Delete(ctx, tr).Return(errors.New("delete error"))

	sut := tax_rate_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	err := sut.Delete(ctx, tr)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delete error")
}

func TestService_QueryByID_Successful(t *testing.T) {
	ctx := context.Background()
	taxRateID := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	expectedTaxRate := tax_rate.New(
		taxRateID,
		"VAT",
		20.0,
		country.MustParse("US"),
		"Standard VAT",
		true,
		mTime,
		mTime,
	)

	mLogger := logger.NewMockLogger(t)

	mStorer := tax_rate.NewMockStorer(t)
	mStorer.EXPECT().QueryByID(ctx, taxRateID).Return(expectedTaxRate, nil)

	sut := tax_rate_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	tr, err := sut.QueryByID(ctx, taxRateID)

	assert.NoError(t, err)
	assert.Equal(t, expectedTaxRate.ID(), tr.ID())
	assert.Equal(t, expectedTaxRate.Name(), tr.Name())
}

func TestService_QueryByID_NotFound(t *testing.T) {
	ctx := context.Background()
	taxRateID := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")

	mLogger := logger.NewMockLogger(t)

	mStorer := tax_rate.NewMockStorer(t)
	mStorer.EXPECT().QueryByID(ctx, taxRateID).Return(tax_rate.TaxRate{}, tax_rate.ErrNotFound)

	sut := tax_rate_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	_, err := sut.QueryByID(ctx, taxRateID)

	assert.Error(t, err)
	assert.ErrorIs(t, err, tax_rate.ErrNotFound)
}

func TestService_Query_Successful(t *testing.T) {
	ctx := context.Background()
	filter := tax_rate.QueryFilter{}
	orderBy := order.By{Field: "name", Direction: "asc"}
	cur := cursor.Cursor{}

	expectedTaxRates := []tax_rate.TaxRate{
		tax_rate.New(
			uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
			"VAT",
			20.0,
			country.MustParse("US"),
			"Standard VAT",
			true,
			time.Time{},
			time.Time{},
		),
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := tax_rate.NewMockStorer(t)
	mStorer.EXPECT().Query(ctx, filter, orderBy, cur).Return(expectedTaxRates, nil)

	sut := tax_rate_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	taxRates, err := sut.Query(ctx, filter, orderBy, cur)

	assert.NoError(t, err)
	assert.Len(t, taxRates, 1)
	assert.Equal(t, expectedTaxRates[0].Name(), taxRates[0].Name())
}

func TestService_Query_Error(t *testing.T) {
	ctx := context.Background()
	filter := tax_rate.QueryFilter{}
	orderBy := order.By{Field: "name", Direction: "asc"}
	cur := cursor.Cursor{}

	mLogger := logger.NewMockLogger(t)

	mStorer := tax_rate.NewMockStorer(t)
	mStorer.EXPECT().Query(ctx, filter, orderBy, cur).Return(nil, errors.New("query error"))

	sut := tax_rate_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	_, err := sut.Query(ctx, filter, orderBy, cur)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "query error")
}
