package user_roles_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Housiadas/cerberus/internal/core/user_roles"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/Housiadas/cerberus/pkg/logger"
)

func TestService_Add_Successful(t *testing.T) {
	ctx := context.Background()
	userID := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")
	roleID := uuid.MustParse("11234567-89ab-7def-0123-456789abcdef")

	mLogger := logger.NewMockLogger(t)
	mStorer := NewMockStorer(t)
	mStorer.EXPECT().Add(ctx, userID, roleID).Return(nil)

	sut := user_roles.NewService(mLogger, mStorer)
	err := sut.Add(ctx, userID, roleID)

	assert.NoError(t, err)
}

func TestService_Add_Error(t *testing.T) {
	ctx := context.Background()
	userID := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")
	roleID := uuid.MustParse("11234567-89ab-7def-0123-456789abcdef")

	mLogger := logger.NewMockLogger(t)
	mStorer := NewMockStorer(t)
	mStorer.EXPECT().Add(ctx, userID, roleID).Return(errors.New("db error"))

	sut := user_roles.NewService(mLogger, mStorer)
	err := sut.Add(ctx, userID, roleID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}

func TestService_Remove_Successful(t *testing.T) {
	ctx := context.Background()
	userID := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")
	roleID := uuid.MustParse("11234567-89ab-7def-0123-456789abcdef")

	mLogger := logger.NewMockLogger(t)
	mStorer := NewMockStorer(t)
	mStorer.EXPECT().Remove(ctx, userID, roleID).Return(nil)

	sut := user_roles.NewService(mLogger, mStorer)
	err := sut.Remove(ctx, userID, roleID)

	assert.NoError(t, err)
}

func TestService_Remove_Error(t *testing.T) {
	ctx := context.Background()
	userID := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")
	roleID := uuid.MustParse("11234567-89ab-7def-0123-456789abcdef")

	mLogger := logger.NewMockLogger(t)
	mStorer := NewMockStorer(t)
	mStorer.EXPECT().Remove(ctx, userID, roleID).Return(errors.New("db error"))

	sut := user_roles.NewService(mLogger, mStorer)
	err := sut.Remove(ctx, userID, roleID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}
