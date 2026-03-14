package plan_service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Housiadas/cerberus/internal/core/domain/currency"
	"github.com/Housiadas/cerberus/internal/core/domain/name"
	"github.com/Housiadas/cerberus/internal/core/domain/plan"
	"github.com/Housiadas/cerberus/internal/core/service/plan_service"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
)

func TestService_Create_Successful(t *testing.T) {
	ctx := context.Background()

	np := plan.NewPlan{
		Name:        name.MustParse("Basic"),
		Description: "Basic plan",
		Interval:    plan.MustParse("monthly"),
		PriceCents:  1000,
		Currency:    currency.MustParse("USD"),
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := plan.NewMockStorer(t)
	mStorer.EXPECT().Create(ctx, mock.AnythingOfType("plan.Plan")).Return(nil)

	mUUID := uuidgen.NewMockGenerator(t)
	mUUID.EXPECT().Generate().Return(uuid.New(), nil)

	sut := plan_service.New(mLogger, mStorer, mUUID)
	pl, err := sut.Create(ctx, np)

	assert.NoError(t, err)
	assert.Equal(t, np.Name, pl.Name())
	assert.Equal(t, np.Description, pl.Description())
	assert.NotZero(t, pl.CreatedAt())
	assert.NotZero(t, pl.UpdatedAt())
}

func TestService_Create_StorerError(t *testing.T) {
	ctx := context.Background()

	np := plan.NewPlan{
		Name:        name.MustParse("Basic"),
		Description: "Basic plan",
		Interval:    plan.MustParse("monthly"),
		PriceCents:  1000,
		Currency:    currency.MustParse("USD"),
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := plan.NewMockStorer(t)
	mStorer.EXPECT().Create(ctx, mock.AnythingOfType("plan.Plan")).Return(errors.New("create error"))

	mUUID := uuidgen.NewMockGenerator(t)
	mUUID.EXPECT().Generate().Return(uuid.New(), nil)

	sut := plan_service.New(mLogger, mStorer, mUUID)
	_, err := sut.Create(ctx, np)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "create error")
}

func TestService_Update_Successful(t *testing.T) {
	ctx := context.Background()
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	existingPlan := plan.New(
		uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		name.MustParse("Basic"),
		"Basic plan",
		plan.MustParse("monthly"),
		1000,
		currency.MustParse("USD"),
		true,
		mTime,
		mTime,
	)

	newName := name.MustParse("Premium")
	up := plan.UpdatePlan{
		Name: &newName,
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := plan.NewMockStorer(t)
	mStorer.EXPECT().Update(ctx, mock.AnythingOfType("plan.Plan")).Return(nil)

	sut := plan_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	pl, err := sut.Update(ctx, existingPlan, up)

	assert.NoError(t, err)
	assert.Equal(t, newName, pl.Name())
	assert.True(t, pl.UpdatedAt().After(mTime) || pl.UpdatedAt().Equal(mTime))
}

func TestService_Update_StorerError(t *testing.T) {
	ctx := context.Background()
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	existingPlan := plan.New(
		uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		name.MustParse("Basic"),
		"Basic plan",
		plan.MustParse("monthly"),
		1000,
		currency.MustParse("USD"),
		true,
		mTime,
		mTime,
	)

	newName := name.MustParse("Premium")
	up := plan.UpdatePlan{
		Name: &newName,
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := plan.NewMockStorer(t)
	mStorer.EXPECT().Update(ctx, mock.AnythingOfType("plan.Plan")).Return(errors.New("update error"))

	sut := plan_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	_, err := sut.Update(ctx, existingPlan, up)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update error")
}

func TestService_Delete_Successful(t *testing.T) {
	ctx := context.Background()

	pl := plan.New(
		uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		name.MustParse("Basic"),
		"Basic plan",
		plan.MustParse("monthly"),
		1000,
		currency.MustParse("USD"),
		true,
		time.Time{},
		time.Time{},
	)

	mLogger := logger.NewMockLogger(t)

	mStorer := plan.NewMockStorer(t)
	mStorer.EXPECT().Delete(ctx, pl).Return(nil)

	sut := plan_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	err := sut.Delete(ctx, pl)

	assert.NoError(t, err)
}

func TestService_Delete_Error(t *testing.T) {
	ctx := context.Background()

	pl := plan.New(
		uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		name.MustParse("Basic"),
		"Basic plan",
		plan.MustParse("monthly"),
		1000,
		currency.MustParse("USD"),
		true,
		time.Time{},
		time.Time{},
	)

	mLogger := logger.NewMockLogger(t)

	mStorer := plan.NewMockStorer(t)
	mStorer.EXPECT().Delete(ctx, pl).Return(errors.New("delete error"))

	sut := plan_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	err := sut.Delete(ctx, pl)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delete error")
}

func TestService_QueryByID_Successful(t *testing.T) {
	ctx := context.Background()
	planID := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	expectedPlan := plan.New(
		planID,
		name.MustParse("Basic"),
		"Basic plan",
		plan.MustParse("monthly"),
		1000,
		currency.MustParse("USD"),
		true,
		mTime,
		mTime,
	)

	mLogger := logger.NewMockLogger(t)

	mStorer := plan.NewMockStorer(t)
	mStorer.EXPECT().QueryByID(ctx, planID).Return(expectedPlan, nil)

	sut := plan_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	pl, err := sut.QueryByID(ctx, planID)

	assert.NoError(t, err)
	assert.Equal(t, expectedPlan.ID(), pl.ID())
	assert.Equal(t, expectedPlan.Name(), pl.Name())
}

func TestService_QueryByID_NotFound(t *testing.T) {
	ctx := context.Background()
	planID := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")

	mLogger := logger.NewMockLogger(t)

	mStorer := plan.NewMockStorer(t)
	mStorer.EXPECT().QueryByID(ctx, planID).Return(plan.Plan{}, plan.ErrNotFound)

	sut := plan_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	_, err := sut.QueryByID(ctx, planID)

	assert.Error(t, err)
	assert.ErrorIs(t, err, plan.ErrNotFound)
}

func TestService_Query_Successful(t *testing.T) {
	ctx := context.Background()
	filter := plan.QueryFilter{}
	orderBy := order.By{Field: "name", Direction: "asc"}
	cur := cursor.Cursor{}

	expectedPlans := []plan.Plan{
		plan.New(
			uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
			name.MustParse("Basic"),
			"Basic plan",
			plan.MustParse("monthly"),
			1000,
			currency.MustParse("USD"),
			true,
			time.Time{},
			time.Time{},
		),
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := plan.NewMockStorer(t)
	mStorer.EXPECT().Query(ctx, filter, orderBy, cur).Return(expectedPlans, nil)

	sut := plan_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	plans, err := sut.Query(ctx, filter, orderBy, cur)

	assert.NoError(t, err)
	assert.Len(t, plans, 1)
	assert.Equal(t, expectedPlans[0].Name(), plans[0].Name())
}

func TestService_Query_Error(t *testing.T) {
	ctx := context.Background()
	filter := plan.QueryFilter{}
	orderBy := order.By{Field: "name", Direction: "asc"}
	cur := cursor.Cursor{}

	mLogger := logger.NewMockLogger(t)

	mStorer := plan.NewMockStorer(t)
	mStorer.EXPECT().Query(ctx, filter, orderBy, cur).Return(nil, errors.New("query error"))

	sut := plan_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	_, err := sut.Query(ctx, filter, orderBy, cur)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "query error")
}
