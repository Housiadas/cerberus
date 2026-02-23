package user_service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/Housiadas/cerberus/internal/core/domain/name"
	"github.com/Housiadas/cerberus/internal/core/domain/user"
	"github.com/Housiadas/cerberus/internal/core/service/user_service"
	"github.com/Housiadas/cerberus/internal/utils/unitest"
	"github.com/Housiadas/cerberus/pkg/clock"
	"github.com/Housiadas/cerberus/pkg/hasher"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
)

func TestService_Authenticate_Successful(t *testing.T) {
	ctx := context.Background()
	email := unitest.MustParseEmail("john@example.com")
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	existingUser := user.User{
		ID:           uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		Name:         name.MustParse("John Doe"),
		Email:        email,
		PasswordHash: []byte("hashed_password"),
		Department:   name.MustParseNull("Engineering"),
		Enabled:      true,
		CreatedAt:    mTime,
		UpdatedAt:    mTime,
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := user.NewMockStorer(t)
	mStorer.EXPECT().QueryByEmail(ctx, email).Return(existingUser, nil)

	mUuidGen := uuidgen.NewMockGenerator(t)
	mClock := clock.NewMockClock(t)

	mHasher := hasher.NewMockHasher(t)
	mHasher.EXPECT().Compare(existingUser.PasswordHash, "secret123").Return(nil)

	sut := user_service.New(mLogger, mStorer, mUuidGen, mClock, mHasher)
	usr, err := sut.Authenticate(ctx, email, "secret123")

	assert.NoError(t, err)
	assert.Equal(t, existingUser.ID, usr.ID)
	assert.Equal(t, existingUser.Name, usr.Name)
	assert.Equal(t, existingUser.Email, usr.Email)
}

func TestService_Authenticate_UserNotFound(t *testing.T) {
	ctx := context.Background()
	email := unitest.MustParseEmail("unknown@example.com")

	mLogger := logger.NewMockLogger(t)

	mStorer := user.NewMockStorer(t)
	mStorer.EXPECT().QueryByEmail(ctx, email).Return(user.User{}, user.ErrNotFound)

	mUuidGen := uuidgen.NewMockGenerator(t)
	mClock := clock.NewMockClock(t)
	mHasher := hasher.NewMockHasher(t)

	sut := user_service.New(mLogger, mStorer, mUuidGen, mClock, mHasher)
	_, err := sut.Authenticate(ctx, email, "secret123")

	assert.Error(t, err)
	assert.ErrorIs(t, err, user.ErrNotFound)
}

func TestService_Authenticate_WrongPassword(t *testing.T) {
	ctx := context.Background()
	email := unitest.MustParseEmail("john@example.com")
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	existingUser := user.User{
		ID:           uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		Name:         name.MustParse("John Doe"),
		Email:        email,
		PasswordHash: []byte("hashed_password"),
		Department:   name.MustParseNull("Engineering"),
		Enabled:      true,
		CreatedAt:    mTime,
		UpdatedAt:    mTime,
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := user.NewMockStorer(t)
	mStorer.EXPECT().QueryByEmail(ctx, email).Return(existingUser, nil)

	mUuidGen := uuidgen.NewMockGenerator(t)
	mClock := clock.NewMockClock(t)

	mHasher := hasher.NewMockHasher(t)
	mHasher.EXPECT().Compare(existingUser.PasswordHash, "wrong_password").Return(errors.New("password mismatch"))

	sut := user_service.New(mLogger, mStorer, mUuidGen, mClock, mHasher)
	_, err := sut.Authenticate(ctx, email, "wrong_password")

	assert.Error(t, err)
	assert.ErrorIs(t, err, user.ErrAuthenticationFailure)
}
