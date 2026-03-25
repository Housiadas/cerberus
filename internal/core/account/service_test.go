package account_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/Housiadas/cerberus/internal/core/account"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestService_Create_Successful(t *testing.T) {
	ctx := context.Background()
	mUuid := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	newAccount := account.NewAccount{
		Name: "Test Account",
	}
	expectedAccount := account.New(
		mUuid,
		"Test Account",
		sql.NullString{},
		mTime,
		mTime,
		nil,
	)

	mLogger := newMocklogger(t)

	mStorer := NewMockStorer(t)
	mStorer.EXPECT().Create(ctx, expectedAccount).Return(nil)

	mUuidGen := newMockgenerator(t)
	mUuidGen.EXPECT().Generate().Return(mUuid, nil)

	mClock := newMockclock(t)
	mClock.EXPECT().Now().Return(mTime)

	sut := account.NewService(mLogger, mStorer, mUuidGen, mClock)
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

	newAccount := account.NewAccount{
		Name: "Test Account",
	}

	mLogger := newMocklogger(t)
	mStorer := NewMockStorer(t)

	mUuidGen := newMockgenerator(t)
	mUuidGen.EXPECT().Generate().Return(mUuid, errors.New("uuid initialization error"))

	mClock := newMockclock(t)

	sut := account.NewService(mLogger, mStorer, mUuidGen, mClock)
	_, err := sut.Create(ctx, newAccount)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "uuid initialization error")
}

func TestService_QueryByID_Successful(t *testing.T) {
	ctx := context.Background()
	mUuid := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	expectedAccount := account.New(
		mUuid,
		"Test Account",
		sql.NullString{String: "cus_test123", Valid: true},
		mTime,
		mTime,
		nil,
	)

	mLogger := newMocklogger(t)

	mStorer := NewMockStorer(t)
	mStorer.EXPECT().QueryByID(ctx, mUuid).Return(expectedAccount, nil)

	mUuidGen := newMockgenerator(t)
	mClock := newMockclock(t)

	sut := account.NewService(mLogger, mStorer, mUuidGen, mClock)
	acc, err := sut.QueryByID(ctx, mUuid)

	assert.NoError(t, err)
	assert.Equal(t, mUuid, acc.ID())
	assert.Equal(t, "Test Account", acc.Name())
}

func TestService_Query_Successful(t *testing.T) {
	ctx := context.Background()
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	expected := []account.Account{
		account.New(
			uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
			"Account 1",
			sql.NullString{},
			mTime,
			mTime,
			nil,
		),
	}

	filter := account.QueryFilter{}
	orderBy := order.NewBy("id", order.ASC)
	cur, _ := cursor.Parse("", "10")

	mLogger := newMocklogger(t)

	mStorer := NewMockStorer(t)
	mStorer.EXPECT().Query(ctx, filter, orderBy, cur).Return(expected, nil)

	mUuidGen := newMockgenerator(t)
	mClock := newMockclock(t)

	sut := account.NewService(mLogger, mStorer, mUuidGen, mClock)
	accs, err := sut.Query(ctx, filter, orderBy, cur)

	assert.NoError(t, err)
	assert.Len(t, accs, 1)
}
