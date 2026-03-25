package user_roles_permissions_service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Housiadas/cerberus/internal/types/name"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/Housiadas/cerberus/internal/core/domain/user_roles_permissions"
	"github.com/Housiadas/cerberus/internal/core/service/user_roles_permissions_service"
	"github.com/Housiadas/cerberus/internal/testutil/unitest"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
)

func TestService_Query_Successful(t *testing.T) {
	ctx := context.Background()
	filter := user_roles_permissions.QueryFilter{}
	orderBy := order.By{Field: "user_name", Direction: "asc"}
	cur := cursor.Cursor{}

	expected := []user_roles_permissions.UserRolesPermissions{
		user_roles_permissions.New(
			uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
			name.MustParse("John Doe"),
			unitest.MustParseEmail("john@example.com"),
			uuid.MustParse("11234567-89ab-7def-0123-456789abcdef"),
			name.MustParse("Admin"),
			nil,
			name.Null{},
		),
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := user_roles_permissions.NewMockStorer(t)
	mStorer.EXPECT().Query(ctx, filter, orderBy, cur).Return(expected, nil)

	sut := user_roles_permissions_service.New(mLogger, mStorer)
	result, err := sut.Query(ctx, filter, orderBy, cur)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, expected[0].UserID(), result[0].UserID())
	assert.Equal(t, expected[0].RoleName(), result[0].RoleName())
}

func TestService_Query_Error(t *testing.T) {
	ctx := context.Background()
	filter := user_roles_permissions.QueryFilter{}
	orderBy := order.By{Field: "user_name", Direction: "asc"}
	cur := cursor.Cursor{}

	mLogger := logger.NewMockLogger(t)

	mStorer := user_roles_permissions.NewMockStorer(t)
	mStorer.EXPECT().Query(ctx, filter, orderBy, cur).Return(nil, errors.New("query error"))

	sut := user_roles_permissions_service.New(mLogger, mStorer)
	_, err := sut.Query(ctx, filter, orderBy, cur)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "query error")
}

func TestService_HasPermission_True(t *testing.T) {
	ctx := context.Background()
	userID := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")

	mLogger := logger.NewMockLogger(t)

	mStorer := user_roles_permissions.NewMockStorer(t)
	mStorer.EXPECT().HasPermission(ctx, userID, "read_users").Return(true, nil)

	sut := user_roles_permissions_service.New(mLogger, mStorer)
	has, err := sut.HasPermission(ctx, userID, "read_users")

	assert.NoError(t, err)
	assert.True(t, has)
}

func TestService_HasPermission_False(t *testing.T) {
	ctx := context.Background()
	userID := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")

	mLogger := logger.NewMockLogger(t)

	mStorer := user_roles_permissions.NewMockStorer(t)
	mStorer.EXPECT().HasPermission(ctx, userID, "delete_users").Return(false, nil)

	sut := user_roles_permissions_service.New(mLogger, mStorer)
	has, err := sut.HasPermission(ctx, userID, "delete_users")

	assert.NoError(t, err)
	assert.False(t, has)
}

func TestService_HasPermission_Error(t *testing.T) {
	ctx := context.Background()
	userID := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")

	mLogger := logger.NewMockLogger(t)

	mStorer := user_roles_permissions.NewMockStorer(t)
	mStorer.EXPECT().HasPermission(ctx, userID, "read_users").Return(false, errors.New("permission error"))

	sut := user_roles_permissions_service.New(mLogger, mStorer)
	_, err := sut.HasPermission(ctx, userID, "read_users")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permission error")
}

func TestService_QueryPermissionsByUserID_Successful(t *testing.T) {
	ctx := context.Background()
	userID := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")

	expected := []user_roles_permissions.Permission{
		{ID: uuid.MustParse("11111111-1111-7111-1111-111111111111"), Name: "user:read"},
		{ID: uuid.MustParse("22222222-2222-7222-2222-222222222222"), Name: "user:write"},
		{ID: uuid.MustParse("33333333-3333-7333-3333-333333333333"), Name: "role:read:all"},
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := user_roles_permissions.NewMockStorer(t)
	mStorer.EXPECT().QueryPermissionsByUserID(ctx, userID).Return(expected, nil)

	sut := user_roles_permissions_service.New(mLogger, mStorer)
	result, err := sut.QueryPermissionsByUserID(ctx, userID)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestService_QueryPermissionsByUserID_Empty(t *testing.T) {
	ctx := context.Background()
	userID := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")

	mLogger := logger.NewMockLogger(t)

	mStorer := user_roles_permissions.NewMockStorer(t)
	mStorer.EXPECT().QueryPermissionsByUserID(ctx, userID).Return([]user_roles_permissions.Permission{}, nil)

	sut := user_roles_permissions_service.New(mLogger, mStorer)
	result, err := sut.QueryPermissionsByUserID(ctx, userID)

	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestService_QueryPermissionsByUserID_Error(t *testing.T) {
	ctx := context.Background()
	userID := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")

	mLogger := logger.NewMockLogger(t)

	mStorer := user_roles_permissions.NewMockStorer(t)
	mStorer.EXPECT().QueryPermissionsByUserID(ctx, userID).Return(nil, errors.New("db error"))

	sut := user_roles_permissions_service.New(mLogger, mStorer)
	_, err := sut.QueryPermissionsByUserID(ctx, userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}
