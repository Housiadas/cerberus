package user_service_test

import (
	"context"
	"errors"
	"net/mail"
	"testing"
	"time"

	"github.com/Housiadas/cerberus/internal/utils/page"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/Housiadas/cerberus/internal/core/domain/name"
	"github.com/Housiadas/cerberus/internal/core/domain/user"
	"github.com/Housiadas/cerberus/internal/core/service/user_service"
	"github.com/Housiadas/cerberus/pkg/clock"
	"github.com/Housiadas/cerberus/pkg/hasher"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
)

func TestService_Query_Successful(t *testing.T) {
	ctx := context.Background()
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)
	filter := user.QueryFilter{}
	orderBy := order.By{Field: "name", Direction: "asc"}
	page := page.Page{}

	email, _ := mail.ParseAddress("john@example.com")
	expectedUsers := []user.User{
		user.New(
			uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
			name.MustParse("John Doe"),
			*email,
			nil,
			name.Null{},
			false,
			mTime,
			mTime,
			nil,
		),
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := user.NewMockStorer(t)
	mStorer.EXPECT().Query(ctx, filter, orderBy, page).Return(expectedUsers, nil)

	mUuidGen := uuidgen.NewMockGenerator(t)
	mClock := clock.NewMockClock(t)
	mHasher := hasher.NewMockHasher(t)

	sut := user_service.New(mLogger, mStorer, mUuidGen, mClock, mHasher)
	users, err := sut.Query(ctx, filter, orderBy, page)

	assert.NoError(t, err)
	assert.Len(t, users, 1)
	assert.Equal(t, expectedUsers[0].ID(), users[0].ID())
}

func TestService_Query_Error(t *testing.T) {
	ctx := context.Background()
	filter := user.QueryFilter{}
	orderBy := order.By{Field: "name", Direction: "asc"}
	page := page.Page{}

	mLogger := logger.NewMockLogger(t)

	mStorer := user.NewMockStorer(t)
	mStorer.EXPECT().Query(ctx, filter, orderBy, page).Return(nil, errors.New("query failed"))

	mUuidGen := uuidgen.NewMockGenerator(t)
	mClock := clock.NewMockClock(t)
	mHasher := hasher.NewMockHasher(t)

	sut := user_service.New(mLogger, mStorer, mUuidGen, mClock, mHasher)
	_, err := sut.Query(ctx, filter, orderBy, page)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "query failed")
}

func TestService_QueryByID_Successful(t *testing.T) {
	ctx := context.Background()
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)
	userID := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")

	email, _ := mail.ParseAddress("john@example.com")
	expectedUser := user.New(
		userID,
		name.MustParse("John Doe"),
		*email,
		nil,
		name.Null{},
		false,
		mTime,
		mTime,
		nil,
	)

	mLogger := logger.NewMockLogger(t)

	mStorer := user.NewMockStorer(t)
	mStorer.EXPECT().QueryByID(ctx, userID).Return(expectedUser, nil)

	mUuidGen := uuidgen.NewMockGenerator(t)
	mClock := clock.NewMockClock(t)
	mHasher := hasher.NewMockHasher(t)

	sut := user_service.New(mLogger, mStorer, mUuidGen, mClock, mHasher)
	usr, err := sut.QueryByID(ctx, userID)

	assert.NoError(t, err)
	assert.Equal(t, expectedUser.ID(), usr.ID())
	assert.Equal(t, expectedUser.Name(), usr.Name())
}

func TestService_QueryByID_NotFound(t *testing.T) {
	ctx := context.Background()
	userID := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")

	mLogger := logger.NewMockLogger(t)

	mStorer := user.NewMockStorer(t)
	mStorer.EXPECT().QueryByID(ctx, userID).Return(user.User{}, user.ErrNotFound)

	mUuidGen := uuidgen.NewMockGenerator(t)
	mClock := clock.NewMockClock(t)
	mHasher := hasher.NewMockHasher(t)

	sut := user_service.New(mLogger, mStorer, mUuidGen, mClock, mHasher)
	_, err := sut.QueryByID(ctx, userID)

	assert.Error(t, err)
	assert.ErrorIs(t, err, user.ErrNotFound)
}

func TestService_QueryByEmail_Successful(t *testing.T) {
	ctx := context.Background()
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)
	email, _ := mail.ParseAddress("john@example.com")

	expectedUser := user.New(
		uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		name.MustParse("John Doe"),
		*email,
		nil,
		name.Null{},
		false,
		mTime,
		mTime,
		nil,
	)

	mLogger := logger.NewMockLogger(t)

	mStorer := user.NewMockStorer(t)
	mStorer.EXPECT().QueryByEmail(ctx, *email).Return(expectedUser, nil)

	mUuidGen := uuidgen.NewMockGenerator(t)
	mClock := clock.NewMockClock(t)
	mHasher := hasher.NewMockHasher(t)

	sut := user_service.New(mLogger, mStorer, mUuidGen, mClock, mHasher)
	usr, err := sut.QueryByEmail(ctx, *email)

	assert.NoError(t, err)
	assert.Equal(t, expectedUser.ID(), usr.ID())
	assert.Equal(t, expectedUser.Email(), usr.Email())
}

func TestService_QueryByEmail_NotFound(t *testing.T) {
	ctx := context.Background()
	email, _ := mail.ParseAddress("unknown@example.com")

	mLogger := logger.NewMockLogger(t)

	mStorer := user.NewMockStorer(t)
	mStorer.EXPECT().QueryByEmail(ctx, *email).Return(user.User{}, user.ErrNotFound)

	mUuidGen := uuidgen.NewMockGenerator(t)
	mClock := clock.NewMockClock(t)
	mHasher := hasher.NewMockHasher(t)

	sut := user_service.New(mLogger, mStorer, mUuidGen, mClock, mHasher)
	_, err := sut.QueryByEmail(ctx, *email)

	assert.Error(t, err)
	assert.ErrorIs(t, err, user.ErrNotFound)
}

func TestService_Count_Successful(t *testing.T) {
	ctx := context.Background()
	filter := user.QueryFilter{}

	mLogger := logger.NewMockLogger(t)

	mStorer := user.NewMockStorer(t)
	mStorer.EXPECT().Count(ctx, filter).Return(5, nil)

	mUuidGen := uuidgen.NewMockGenerator(t)
	mClock := clock.NewMockClock(t)
	mHasher := hasher.NewMockHasher(t)

	sut := user_service.New(mLogger, mStorer, mUuidGen, mClock, mHasher)
	count, err := sut.Count(ctx, filter)

	assert.NoError(t, err)
	assert.Equal(t, 5, count)
}

func TestService_Count_Error(t *testing.T) {
	ctx := context.Background()
	filter := user.QueryFilter{}

	mLogger := logger.NewMockLogger(t)

	mStorer := user.NewMockStorer(t)
	mStorer.EXPECT().Count(ctx, filter).Return(0, errors.New("count failed"))

	mUuidGen := uuidgen.NewMockGenerator(t)
	mClock := clock.NewMockClock(t)
	mHasher := hasher.NewMockHasher(t)

	sut := user_service.New(mLogger, mStorer, mUuidGen, mClock, mHasher)
	_, err := sut.Count(ctx, filter)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "count failed")
}
