package user_service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/Housiadas/cerberus/internal/core/domain/name"
	"github.com/Housiadas/cerberus/internal/core/domain/password"
	"github.com/Housiadas/cerberus/internal/core/domain/user"
	"github.com/Housiadas/cerberus/internal/core/service/user_service"
	"github.com/Housiadas/cerberus/internal/utils/unitest"
	"github.com/Housiadas/cerberus/pkg/clock"
	"github.com/Housiadas/cerberus/pkg/hasher"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
)

func TestService_Update_AllFields(t *testing.T) {
	ctx := context.Background()
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)
	updatedTime := time.Date(2026, 1, 2, 10, 30, 0, 0, time.UTC)

	existingUser := user.User{
		ID:           uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		Name:         name.MustParse("John Doe"),
		Email:        unitest.MustParseEmail("john@example.com"),
		PasswordHash: []byte("old_hash"),
		Department:   name.MustParseNull("Engineering"),
		Enabled:      true,
		CreatedAt:    mTime,
		UpdatedAt:    mTime,
	}

	newName := name.MustParse("Jane Doe")
	newEmail := unitest.MustParseEmail("jane@example.com")
	newPassword := password.MustParse("newpassword123")
	newDepartment := name.MustParseNull("Marketing")
	newEnabled := false

	uu := user.UpdateUser{
		Name:       &newName,
		Email:      &newEmail,
		Password:   &newPassword,
		Department: &newDepartment,
		Enabled:    &newEnabled,
	}

	expectedUser := user.User{
		ID:           existingUser.ID,
		Name:         newName,
		Email:        newEmail,
		PasswordHash: []byte("new_hash"),
		Department:   newDepartment,
		Enabled:      false,
		CreatedAt:    mTime,
		UpdatedAt:    updatedTime,
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := user.NewMockStorer(t)
	mStorer.EXPECT().Update(ctx, expectedUser).Return(nil)

	mUuidGen := uuidgen.NewMockGenerator(t)

	mClock := clock.NewMockClock(t)
	mClock.EXPECT().Now().Return(updatedTime)

	mHasher := hasher.NewMockHasher(t)
	mHasher.EXPECT().Hash(newPassword.String()).Return([]byte("new_hash"), nil)

	sut := user_service.New(mLogger, mStorer, mUuidGen, mClock, mHasher)
	usr, err := sut.Update(ctx, existingUser, uu)

	assert.NoError(t, err)
	assert.Equal(t, newName, usr.Name)
	assert.Equal(t, newEmail, usr.Email)
	assert.Equal(t, newDepartment, usr.Department)
	assert.Equal(t, false, usr.Enabled)
	assert.Equal(t, updatedTime, usr.UpdatedAt)
}

func TestService_Update_PartialFields(t *testing.T) {
	ctx := context.Background()
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)
	updatedTime := time.Date(2026, 1, 2, 10, 30, 0, 0, time.UTC)

	existingUser := user.User{
		ID:           uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		Name:         name.MustParse("John Doe"),
		Email:        unitest.MustParseEmail("john@example.com"),
		PasswordHash: []byte("old_hash"),
		Department:   name.MustParseNull("Engineering"),
		Enabled:      true,
		CreatedAt:    mTime,
		UpdatedAt:    mTime,
	}

	newName := name.MustParse("Jane Doe")
	uu := user.UpdateUser{
		Name: &newName,
	}

	expectedUser := user.User{
		ID:           existingUser.ID,
		Name:         newName,
		Email:        existingUser.Email,
		PasswordHash: existingUser.PasswordHash,
		Department:   existingUser.Department,
		Enabled:      true,
		CreatedAt:    mTime,
		UpdatedAt:    updatedTime,
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := user.NewMockStorer(t)
	mStorer.EXPECT().Update(ctx, expectedUser).Return(nil)

	mUuidGen := uuidgen.NewMockGenerator(t)

	mClock := clock.NewMockClock(t)
	mClock.EXPECT().Now().Return(updatedTime)

	mHasher := hasher.NewMockHasher(t)

	sut := user_service.New(mLogger, mStorer, mUuidGen, mClock, mHasher)
	usr, err := sut.Update(ctx, existingUser, uu)

	assert.NoError(t, err)
	assert.Equal(t, newName, usr.Name)
	assert.Equal(t, existingUser.Email, usr.Email)
}

func TestService_Update_HasherError(t *testing.T) {
	ctx := context.Background()
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	existingUser := user.User{
		ID:           uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		Name:         name.MustParse("John Doe"),
		Email:        unitest.MustParseEmail("john@example.com"),
		PasswordHash: []byte("old_hash"),
		Department:   name.MustParseNull("Engineering"),
		Enabled:      true,
		CreatedAt:    mTime,
		UpdatedAt:    mTime,
	}

	newPassword := password.MustParse("newpassword123")
	uu := user.UpdateUser{
		Password: &newPassword,
	}

	mLogger := logger.NewMockLogger(t)
	mStorer := user.NewMockStorer(t)
	mUuidGen := uuidgen.NewMockGenerator(t)
	mClock := clock.NewMockClock(t)

	mHasher := hasher.NewMockHasher(t)
	mHasher.EXPECT().Hash(newPassword.String()).Return(nil, errors.New("hash error"))

	sut := user_service.New(mLogger, mStorer, mUuidGen, mClock, mHasher)
	_, err := sut.Update(ctx, existingUser, uu)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "hash error")
}

func TestService_Update_StorerError(t *testing.T) {
	ctx := context.Background()
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)
	updatedTime := time.Date(2026, 1, 2, 10, 30, 0, 0, time.UTC)

	existingUser := user.User{
		ID:           uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		Name:         name.MustParse("John Doe"),
		Email:        unitest.MustParseEmail("john@example.com"),
		PasswordHash: []byte("old_hash"),
		Department:   name.MustParseNull("Engineering"),
		Enabled:      true,
		CreatedAt:    mTime,
		UpdatedAt:    mTime,
	}

	newName := name.MustParse("Jane Doe")
	uu := user.UpdateUser{
		Name: &newName,
	}

	expectedUser := user.User{
		ID:           existingUser.ID,
		Name:         newName,
		Email:        existingUser.Email,
		PasswordHash: existingUser.PasswordHash,
		Department:   existingUser.Department,
		Enabled:      true,
		CreatedAt:    mTime,
		UpdatedAt:    updatedTime,
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := user.NewMockStorer(t)
	mStorer.EXPECT().Update(ctx, expectedUser).Return(errors.New("update failed"))

	mUuidGen := uuidgen.NewMockGenerator(t)

	mClock := clock.NewMockClock(t)
	mClock.EXPECT().Now().Return(updatedTime)

	mHasher := hasher.NewMockHasher(t)

	sut := user_service.New(mLogger, mStorer, mUuidGen, mClock, mHasher)
	_, err := sut.Update(ctx, existingUser, uu)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update failed")
}
