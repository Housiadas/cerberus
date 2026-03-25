package role_permissions_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Housiadas/cerberus/internal/core/role_permissions"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/Housiadas/cerberus/pkg/logger"
)

func TestService_Add_Successful(t *testing.T) {
	ctx := context.Background()
	roleID := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")
	permID := uuid.MustParse("11234567-89ab-7def-0123-456789abcdef")

	mLogger := logger.NewMockLogger(t)
	mStorer := NewMockStorer(t)
	mStorer.EXPECT().Add(ctx, roleID, permID).Return(nil)

	sut := role_permissions.NewService(mLogger, mStorer)
	err := sut.Add(ctx, roleID, permID)

	assert.NoError(t, err)
}

func TestService_Add_Error(t *testing.T) {
	ctx := context.Background()
	roleID := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")
	permID := uuid.MustParse("11234567-89ab-7def-0123-456789abcdef")

	mLogger := logger.NewMockLogger(t)
	mStorer := NewMockStorer(t)
	mStorer.EXPECT().Add(ctx, roleID, permID).Return(errors.New("db error"))

	sut := role_permissions.NewService(mLogger, mStorer)
	err := sut.Add(ctx, roleID, permID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}

func TestService_Remove_Successful(t *testing.T) {
	ctx := context.Background()
	roleID := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")
	permID := uuid.MustParse("11234567-89ab-7def-0123-456789abcdef")

	mLogger := logger.NewMockLogger(t)
	mStorer := NewMockStorer(t)
	mStorer.EXPECT().Remove(ctx, roleID, permID).Return(nil)

	sut := role_permissions.NewService(mLogger, mStorer)
	err := sut.Remove(ctx, roleID, permID)

	assert.NoError(t, err)
}

func TestService_Remove_Error(t *testing.T) {
	ctx := context.Background()
	roleID := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")
	permID := uuid.MustParse("11234567-89ab-7def-0123-456789abcdef")

	mLogger := logger.NewMockLogger(t)
	mStorer := NewMockStorer(t)
	mStorer.EXPECT().Remove(ctx, roleID, permID).Return(errors.New("db error"))

	sut := role_permissions.NewService(mLogger, mStorer)
	err := sut.Remove(ctx, roleID, permID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}
