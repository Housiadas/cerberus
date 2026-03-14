package account_service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Housiadas/cerberus/internal/core/domain/account"
	"github.com/Housiadas/cerberus/internal/core/domain/name"
	"github.com/Housiadas/cerberus/internal/core/service/account_service"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
)

func TestService_Create_Successful(t *testing.T) {
	ctx := context.Background()

	na := account.NewAccount{
		Name: name.MustParse("Test"),
		Type: account.MustParse("personal"),
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := account.NewMockStorer(t)
	mStorer.EXPECT().Create(ctx, mock.AnythingOfType("account.Account")).Return(nil)

	mUUID := uuidgen.NewMockGenerator(t)
	mUUID.EXPECT().Generate().Return(uuid.New(), nil)

	sut := account_service.New(mLogger, mStorer, mUUID)
	acct, err := sut.Create(ctx, na)

	assert.NoError(t, err)
	assert.Equal(t, na.Name, acct.Name())
	assert.Equal(t, na.Type, acct.Type())
	assert.NotZero(t, acct.CreatedAt())
	assert.NotZero(t, acct.UpdatedAt())
}

func TestService_Create_StorerError(t *testing.T) {
	ctx := context.Background()

	na := account.NewAccount{
		Name: name.MustParse("Test"),
		Type: account.MustParse("personal"),
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := account.NewMockStorer(t)
	mStorer.EXPECT().Create(ctx, mock.AnythingOfType("account.Account")).Return(errors.New("create error"))

	mUUID := uuidgen.NewMockGenerator(t)
	mUUID.EXPECT().Generate().Return(uuid.New(), nil)

	sut := account_service.New(mLogger, mStorer, mUUID)
	_, err := sut.Create(ctx, na)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "create error")
}

func TestService_Update_Successful(t *testing.T) {
	ctx := context.Background()
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	existingAccount := account.New(
		uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		name.MustParse("Test"),
		account.MustParse("personal"),
		true,
		mTime,
		mTime,
		nil,
	)

	newName := name.MustParse("Updated")
	ua := account.UpdateAccount{
		Name: &newName,
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := account.NewMockStorer(t)
	mStorer.EXPECT().Update(ctx, mock.AnythingOfType("account.Account")).Return(nil)

	sut := account_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	acct, err := sut.Update(ctx, existingAccount, ua)

	assert.NoError(t, err)
	assert.Equal(t, newName, acct.Name())
	assert.True(t, acct.UpdatedAt().After(mTime) || acct.UpdatedAt().Equal(mTime))
}

func TestService_Update_StorerError(t *testing.T) {
	ctx := context.Background()
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	existingAccount := account.New(
		uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		name.MustParse("Test"),
		account.MustParse("personal"),
		true,
		mTime,
		mTime,
		nil,
	)

	newName := name.MustParse("Updated")
	ua := account.UpdateAccount{
		Name: &newName,
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := account.NewMockStorer(t)
	mStorer.EXPECT().Update(ctx, mock.AnythingOfType("account.Account")).Return(errors.New("update error"))

	sut := account_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	_, err := sut.Update(ctx, existingAccount, ua)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update error")
}

func TestService_Delete_Successful(t *testing.T) {
	ctx := context.Background()

	acct := account.New(
		uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		name.MustParse("Test"),
		account.MustParse("personal"),
		true,
		time.Time{},
		time.Time{},
		nil,
	)

	mLogger := logger.NewMockLogger(t)

	mStorer := account.NewMockStorer(t)
	mStorer.EXPECT().Delete(ctx, acct).Return(nil)

	sut := account_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	err := sut.Delete(ctx, acct)

	assert.NoError(t, err)
}

func TestService_Delete_Error(t *testing.T) {
	ctx := context.Background()

	acct := account.New(
		uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		name.MustParse("Test"),
		account.MustParse("personal"),
		true,
		time.Time{},
		time.Time{},
		nil,
	)

	mLogger := logger.NewMockLogger(t)

	mStorer := account.NewMockStorer(t)
	mStorer.EXPECT().Delete(ctx, acct).Return(errors.New("delete error"))

	sut := account_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	err := sut.Delete(ctx, acct)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delete error")
}

func TestService_QueryByID_Successful(t *testing.T) {
	ctx := context.Background()
	accountID := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	expectedAccount := account.New(
		accountID,
		name.MustParse("Test"),
		account.MustParse("personal"),
		true,
		mTime,
		mTime,
		nil,
	)

	mLogger := logger.NewMockLogger(t)

	mStorer := account.NewMockStorer(t)
	mStorer.EXPECT().QueryByID(ctx, accountID).Return(expectedAccount, nil)

	sut := account_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	acct, err := sut.QueryByID(ctx, accountID)

	assert.NoError(t, err)
	assert.Equal(t, expectedAccount.ID(), acct.ID())
	assert.Equal(t, expectedAccount.Name(), acct.Name())
}

func TestService_QueryByID_NotFound(t *testing.T) {
	ctx := context.Background()
	accountID := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")

	mLogger := logger.NewMockLogger(t)

	mStorer := account.NewMockStorer(t)
	mStorer.EXPECT().QueryByID(ctx, accountID).Return(account.Account{}, account.ErrNotFound)

	sut := account_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	_, err := sut.QueryByID(ctx, accountID)

	assert.Error(t, err)
	assert.ErrorIs(t, err, account.ErrNotFound)
}

func TestService_Query_Successful(t *testing.T) {
	ctx := context.Background()
	filter := account.QueryFilter{}
	orderBy := order.By{Field: "name", Direction: "asc"}
	cur := cursor.Cursor{}

	expectedAccounts := []account.Account{
		account.New(
			uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
			name.MustParse("Test"),
			account.MustParse("personal"),
			true,
			time.Time{},
			time.Time{},
			nil,
		),
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := account.NewMockStorer(t)
	mStorer.EXPECT().Query(ctx, filter, orderBy, cur).Return(expectedAccounts, nil)

	sut := account_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	accounts, err := sut.Query(ctx, filter, orderBy, cur)

	assert.NoError(t, err)
	assert.Len(t, accounts, 1)
	assert.Equal(t, expectedAccounts[0].Name(), accounts[0].Name())
}

func TestService_Query_Error(t *testing.T) {
	ctx := context.Background()
	filter := account.QueryFilter{}
	orderBy := order.By{Field: "name", Direction: "asc"}
	cur := cursor.Cursor{}

	mLogger := logger.NewMockLogger(t)

	mStorer := account.NewMockStorer(t)
	mStorer.EXPECT().Query(ctx, filter, orderBy, cur).Return(nil, errors.New("query error"))

	sut := account_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	_, err := sut.Query(ctx, filter, orderBy, cur)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "query error")
}
