package user_service_test

import (
	"context"
	"errors"
	"net/mail"
	"testing"
	"time"

	"github.com/Housiadas/cerberus/internal/core/user/user_service"
	"github.com/Housiadas/cerberus/internal/types/name"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/Housiadas/cerberus/internal/core/user"
	"github.com/Housiadas/cerberus/pkg/clock"
	"github.com/Housiadas/cerberus/pkg/hasher"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
)

func TestService_Delete_Successful(t *testing.T) {
	ctx := context.Background()
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	email, _ := mail.ParseAddress("john@example.com")
	usr := user.New(
		uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		name.MustParse("John Doe"),
		*email,
		[]byte("hashed_password"),
		name.MustParseNull("Engineering"),
		true,
		nil,
		mTime,
		mTime,
		nil,
	)

	mLogger := logger.NewMockLogger(t)

	mStorer := user.NewMockStorer(t)
	mStorer.EXPECT().Delete(ctx, usr).Return(nil)

	mUuidGen := uuidgen.NewMockGenerator(t)
	mClock := clock.NewMockClock(t)
	mHasher := hasher.NewMockHasher(t)

	sut := user_service.New(mLogger, mStorer, mUuidGen, mClock, mHasher)
	err := sut.Delete(ctx, usr)

	assert.NoError(t, err)
}

func TestService_Delete_Error(t *testing.T) {
	ctx := context.Background()
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	email, _ := mail.ParseAddress("john@example.com")
	usr := user.New(
		uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		name.MustParse("John Doe"),
		*email,
		[]byte("hashed_password"),
		name.MustParseNull("Engineering"),
		true,
		nil,
		mTime,
		mTime,
		nil,
	)

	mLogger := logger.NewMockLogger(t)

	mStorer := user.NewMockStorer(t)
	mStorer.EXPECT().Delete(ctx, usr).Return(errors.New("delete failed"))

	mUuidGen := uuidgen.NewMockGenerator(t)
	mClock := clock.NewMockClock(t)
	mHasher := hasher.NewMockHasher(t)

	sut := user_service.New(mLogger, mStorer, mUuidGen, mClock, mHasher)
	err := sut.Delete(ctx, usr)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delete failed")
}
