package account_service_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	account2 "github.com/Housiadas/cerberus/internal/core/account"
	"github.com/Housiadas/cerberus/internal/core/account/account_service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/Housiadas/cerberus/pkg/clock"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
)

func TestService_QueryByID_Successful(t *testing.T) {
	ctx := context.Background()
	mUuid := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	expectedAccount := account2.New(
		mUuid,
		"Test Account",
		sql.NullString{String: "cus_test123", Valid: true},
		mTime,
		mTime,
		nil,
	)

	mLogger := logger.NewMockLogger(t)

	mStorer := account2.NewMockStorer(t)
	mStorer.EXPECT().QueryByID(ctx, mUuid).Return(expectedAccount, nil)

	mUuidGen := uuidgen.NewMockGenerator(t)
	mClock := clock.NewMockClock(t)

	sut := account_service.New(mLogger, mStorer, mUuidGen, mClock)
	acc, err := sut.QueryByID(ctx, mUuid)

	assert.NoError(t, err)
	assert.Equal(t, mUuid, acc.ID())
	assert.Equal(t, "Test Account", acc.Name())
}

func TestService_Query_Successful(t *testing.T) {
	ctx := context.Background()
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	expected := []account2.Account{
		account2.New(
			uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
			"Account 1",
			sql.NullString{},
			mTime,
			mTime,
			nil,
		),
	}

	filter := account2.QueryFilter{}
	orderBy := order.NewBy("id", order.ASC)
	cur, _ := cursor.Parse("", "10")

	mLogger := logger.NewMockLogger(t)

	mStorer := account2.NewMockStorer(t)
	mStorer.EXPECT().Query(ctx, filter, orderBy, cur).Return(expected, nil)

	mUuidGen := uuidgen.NewMockGenerator(t)
	mClock := clock.NewMockClock(t)

	sut := account_service.New(mLogger, mStorer, mUuidGen, mClock)
	accs, err := sut.Query(ctx, filter, orderBy, cur)

	assert.NoError(t, err)
	assert.Len(t, accs, 1)
}
