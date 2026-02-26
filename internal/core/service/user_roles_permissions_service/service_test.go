package user_roles_permissions_service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Housiadas/cerberus/internal/utils/page"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/Housiadas/cerberus/internal/core/domain/name"
	"github.com/Housiadas/cerberus/internal/core/domain/user_roles_permissions"
	"github.com/Housiadas/cerberus/internal/core/service/user_roles_permissions_service"
	"github.com/Housiadas/cerberus/internal/utils/unitest"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
)

func TestService_Query_Successful(t *testing.T) {
	ctx := context.Background()
	filter := user_roles_permissions.QueryFilter{}
	orderBy := order.By{Field: "user_name", Direction: "asc"}
	page := page.Page{}

	expected := []user_roles_permissions.UserRolesPermissions{
		{
			UserID:    uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
			UserName:  name.MustParse("John Doe"),
			UserEmail: unitest.MustParseEmail("john@example.com"),
			RoleID:    uuid.MustParse("11234567-89ab-7def-0123-456789abcdef"),
			RoleName:  name.MustParse("Admin"),
		},
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := user_roles_permissions.NewMockStorer(t)
	mStorer.EXPECT().Query(ctx, filter, orderBy, page).Return(expected, nil)

	sut := user_roles_permissions_service.New(mLogger, mStorer)
	result, err := sut.Query(ctx, filter, orderBy, page)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, expected[0].UserID, result[0].UserID)
	assert.Equal(t, expected[0].RoleName, result[0].RoleName)
}

func TestService_Query_Error(t *testing.T) {
	ctx := context.Background()
	filter := user_roles_permissions.QueryFilter{}
	orderBy := order.By{Field: "user_name", Direction: "asc"}
	page := page.Page{}

	mLogger := logger.NewMockLogger(t)

	mStorer := user_roles_permissions.NewMockStorer(t)
	mStorer.EXPECT().Query(ctx, filter, orderBy, page).Return(nil, errors.New("query error"))

	sut := user_roles_permissions_service.New(mLogger, mStorer)
	_, err := sut.Query(ctx, filter, orderBy, page)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "query error")
}

func TestService_Count_Successful(t *testing.T) {
	ctx := context.Background()
	filter := user_roles_permissions.QueryFilter{}

	mLogger := logger.NewMockLogger(t)

	mStorer := user_roles_permissions.NewMockStorer(t)
	mStorer.EXPECT().Count(ctx, filter).Return(7, nil)

	sut := user_roles_permissions_service.New(mLogger, mStorer)
	count, err := sut.Count(ctx, filter)

	assert.NoError(t, err)
	assert.Equal(t, 7, count)
}

func TestService_Count_Error(t *testing.T) {
	ctx := context.Background()
	filter := user_roles_permissions.QueryFilter{}

	mLogger := logger.NewMockLogger(t)

	mStorer := user_roles_permissions.NewMockStorer(t)
	mStorer.EXPECT().Count(ctx, filter).Return(0, errors.New("count error"))

	sut := user_roles_permissions_service.New(mLogger, mStorer)
	_, err := sut.Count(ctx, filter)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "count error")
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
