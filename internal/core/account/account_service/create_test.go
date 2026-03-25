package account_service_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	account2 "github.com/Housiadas/cerberus/internal/core/account"
	"github.com/Housiadas/cerberus/internal/core/account/account_service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/Housiadas/cerberus/pkg/clock"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
)

func TestService_Create_Successful(t *testing.T) {
	ctx := context.Background()
	mUuid := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	newAccount := account2.NewAccount{
		Name: "Test Account",
	}
	expectedAccount := account2.New(
		mUuid,
		"Test Account",
		sql.NullString{},
		mTime,
		mTime,
		nil,
	)

	mLogger := logger.NewMockLogger(t)

	mStorer := account2.NewMockStorer(t)
	mStorer.EXPECT().Create(ctx, expectedAccount).Return(nil)

	mUuidGen := uuidgen.NewMockGenerator(t)
	mUuidGen.EXPECT().Generate().Return(mUuid, nil)

	mClock := clock.NewMockClock(t)
	mClock.EXPECT().Now().Return(mTime)

	sut := account_service.New(mLogger, mStorer, mUuidGen, mClock)
	acc, err := sut.Create(ctx, newAccount)

	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, acc.ID())
	assert.Equal(t, newAccount.Name, acc.Name())
	assert.NotZero(t, acc.CreatedAt())
	assert.NotZero(t, acc.UpdatedAt())
}

func TestService_Create_Uuid_Error(t *testing.T) {
	ctx := context.Background()
	mUuid := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")

	newAccount := account2.NewAccount{
		Name: "Test Account",
	}

	mLogger := logger.NewMockLogger(t)
	mStorer := account2.NewMockStorer(t)

	mUuidGen := uuidgen.NewMockGenerator(t)
	mUuidGen.EXPECT().Generate().Return(mUuid, errors.New("uuid initialization error"))

	mClock := clock.NewMockClock(t)

	sut := account_service.New(mLogger, mStorer, mUuidGen, mClock)
	_, err := sut.Create(ctx, newAccount)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "uuid initialization error")
}
