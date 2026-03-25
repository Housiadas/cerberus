package user_service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Housiadas/cerberus/internal/core/user/user_service"
	"github.com/Housiadas/cerberus/internal/types/name"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/Housiadas/cerberus/internal/core/user"
	"github.com/Housiadas/cerberus/internal/testutil/unitest"
	"github.com/Housiadas/cerberus/pkg/clock"
	"github.com/Housiadas/cerberus/pkg/hasher"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
)

func TestService_VerifyPassword_Successful(t *testing.T) {
	ctx := context.Background()
	_ = ctx
	email := unitest.MustParseEmail("john@example.com")
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	existingUser := user.New(
		uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		name.MustParse("John Doe"),
		email,
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
	mUuidGen := uuidgen.NewMockGenerator(t)
	mClock := clock.NewMockClock(t)

	mHasher := hasher.NewMockHasher(t)
	mHasher.EXPECT().Compare(existingUser.PasswordHash(), "Secret123!@#").Return(nil)

	sut := user_service.New(mLogger, mStorer, mUuidGen, mClock, mHasher)
	err := sut.VerifyPassword(existingUser, "Secret123!@#")

	assert.NoError(t, err)
}

func TestService_VerifyPassword_WrongPassword(t *testing.T) {
	ctx := context.Background()
	_ = ctx
	email := unitest.MustParseEmail("john@example.com")
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	existingUser := user.New(
		uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		name.MustParse("John Doe"),
		email,
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
	mUuidGen := uuidgen.NewMockGenerator(t)
	mClock := clock.NewMockClock(t)

	mHasher := hasher.NewMockHasher(t)
	mHasher.EXPECT().Compare(existingUser.PasswordHash(), "wrongPassword").Return(errors.New("password mismatch"))

	sut := user_service.New(mLogger, mStorer, mUuidGen, mClock, mHasher)
	err := sut.VerifyPassword(existingUser, "wrongPassword")

	assert.Error(t, err)
	assert.ErrorIs(t, err, user.ErrAuthenticationFailure)
}
