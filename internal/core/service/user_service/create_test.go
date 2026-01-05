package user_service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/Housiadas/cerberus/internal/common/unitest"
	"github.com/Housiadas/cerberus/internal/core/domain/name"
	"github.com/Housiadas/cerberus/internal/core/domain/password"
	"github.com/Housiadas/cerberus/internal/core/domain/user"
	"github.com/Housiadas/cerberus/internal/core/service/user_service"
	"github.com/Housiadas/cerberus/pkg/clock"
	"github.com/Housiadas/cerberus/pkg/hasher"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
)

func TestService_Create_Successful(t *testing.T) {
	ctx := context.Background()
	mUuid := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	newUser := user.NewUser{
		Name:       name.MustParse("John Doe"),
		Email:      unitest.MustParseEmail("john@example.com"),
		Password:   password.MustParse("password123"),
		Department: name.MustParseNull("Engineering"),
	}
	expectedUser := user.User{
		ID:           mUuid,
		Name:         name.MustParse("John Doe"),
		Email:        unitest.MustParseEmail("john@example.com"),
		PasswordHash: []byte("password123"),
		Department:   name.MustParseNull("Engineering"),
		Enabled:      true,
		CreatedAt:    mTime,
		UpdatedAt:    mTime,
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := user.NewMockStorer(t)
	mStorer.EXPECT().Create(ctx, expectedUser).Return(nil)

	mUuidGen := uuidgen.NewMockGenerator(t)
	mUuidGen.EXPECT().Generate().Return(mUuid, nil)

	mClock := clock.NewMockClock(t)
	mClock.EXPECT().Now().Return(mTime)

	mHasher := hasher.NewMockHasher(t)
	mHasher.EXPECT().Hash(newUser.Password.String()).Return(expectedUser.PasswordHash, nil)

	sut := user_service.New(mLogger, mStorer, mUuidGen, mClock, mHasher)
	usr, err := sut.Create(ctx, newUser)

	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, usr.ID)
	assert.Equal(t, newUser.Name, usr.Name)
	assert.Equal(t, newUser.Email, usr.Email)
	assert.Equal(t, newUser.Department, usr.Department)
	assert.NotZero(t, usr.CreatedAt)
	assert.NotZero(t, usr.UpdatedAt)
}

func TestService_Create_Uuid_Error(t *testing.T) {
	ctx := context.Background()
	mUuid := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")

	newUser := user.NewUser{
		Name:       name.MustParse("John Doe"),
		Email:      unitest.MustParseEmail("john@example.com"),
		Password:   password.MustParse("password123"),
		Department: name.MustParseNull("Engineering"),
	}

	mLogger := logger.NewMockLogger(t)
	mStorer := user.NewMockStorer(t)

	mUuidGen := uuidgen.NewMockGenerator(t)
	mUuidGen.EXPECT().Generate().Return(mUuid, errors.New("uuid initialization error"))

	mClock := clock.NewMockClock(t)
	mHasher := hasher.NewMockHasher(t)

	sut := user_service.New(mLogger, mStorer, mUuidGen, mClock, mHasher)
	_, err := sut.Create(ctx, newUser)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "uuid initialization error")
}

func TestService_Create_Hasher_Error(t *testing.T) {
	ctx := context.Background()
	mUuid := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")

	newUser := user.NewUser{
		Name:       name.MustParse("John Doe"),
		Email:      unitest.MustParseEmail("john@example.com"),
		Password:   password.MustParse("password123"),
		Department: name.MustParseNull("Engineering"),
	}

	mLogger := logger.NewMockLogger(t)
	mStorer := user.NewMockStorer(t)

	mUuidGen := uuidgen.NewMockGenerator(t)
	mUuidGen.EXPECT().Generate().Return(mUuid, nil)

	mClock := clock.NewMockClock(t)
	mHasher := hasher.NewMockHasher(t)
	mHasher.EXPECT().
		Hash(newUser.Password.String()).
		Return(nil, errors.New("hash initialization error"))

	sut := user_service.New(mLogger, mStorer, mUuidGen, mClock, mHasher)
	_, err := sut.Create(ctx, newUser)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "hash initialization error")
}
