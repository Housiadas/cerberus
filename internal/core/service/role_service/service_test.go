package role_service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Housiadas/cerberus/internal/types/name"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Housiadas/cerberus/internal/core/domain/role"
	"github.com/Housiadas/cerberus/internal/core/service/role_service"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
)

func TestService_Create_Successful(t *testing.T) {
	ctx := context.Background()

	nr := role.NewRole{
		Name: name.MustParse("Admin"),
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := role.NewMockStorer(t)
	mStorer.EXPECT().Create(ctx, mock.AnythingOfType("role.Role")).Return(nil)

	mUUID := uuidgen.NewMockGenerator(t)
	mUUID.EXPECT().Generate().Return(uuid.New(), nil)

	sut := role_service.New(mLogger, mStorer, mUUID)
	rl, err := sut.Create(ctx, nr)

	assert.NoError(t, err)
	assert.Equal(t, nr.Name, rl.Name())
	assert.NotZero(t, rl.CreatedAt())
	assert.NotZero(t, rl.UpdatedAt())
}

func TestService_Create_StorerError(t *testing.T) {
	ctx := context.Background()

	nr := role.NewRole{
		Name: name.MustParse("Admin"),
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := role.NewMockStorer(t)
	mStorer.EXPECT().Create(ctx, mock.AnythingOfType("role.Role")).Return(errors.New("create error"))

	mUUID := uuidgen.NewMockGenerator(t)
	mUUID.EXPECT().Generate().Return(uuid.New(), nil)

	sut := role_service.New(mLogger, mStorer, mUUID)
	_, err := sut.Create(ctx, nr)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "create error")
}

func TestService_Update_Successful(t *testing.T) {
	ctx := context.Background()
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	existingRole := role.New(
		uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		name.MustParse("Admin"),
		mTime,
		mTime,
		nil,
	)

	newName := name.MustParse("SuperAdmin")
	ur := role.UpdateRole{
		Name: &newName,
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := role.NewMockStorer(t)
	mStorer.EXPECT().Update(ctx, mock.AnythingOfType("role.Role")).Return(nil)

	sut := role_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	rl, err := sut.Update(ctx, existingRole, ur)

	assert.NoError(t, err)
	assert.Equal(t, newName, rl.Name())
	assert.True(t, rl.UpdatedAt().After(mTime) || rl.UpdatedAt().Equal(mTime))
}

func TestService_Update_StorerError(t *testing.T) {
	ctx := context.Background()
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	existingRole := role.New(
		uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		name.MustParse("Admin"),
		mTime,
		mTime,
		nil,
	)

	newName := name.MustParse("SuperAdmin")
	ur := role.UpdateRole{
		Name: &newName,
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := role.NewMockStorer(t)
	mStorer.EXPECT().Update(ctx, mock.AnythingOfType("role.Role")).Return(errors.New("update error"))

	sut := role_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	_, err := sut.Update(ctx, existingRole, ur)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update error")
}

func TestService_Delete_Successful(t *testing.T) {
	ctx := context.Background()

	rl := role.New(
		uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		name.MustParse("Admin"),
		time.Time{},
		time.Time{},
		nil,
	)

	mLogger := logger.NewMockLogger(t)

	mStorer := role.NewMockStorer(t)
	mStorer.EXPECT().Delete(ctx, rl).Return(nil)

	sut := role_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	err := sut.Delete(ctx, rl)

	assert.NoError(t, err)
}

func TestService_Delete_Error(t *testing.T) {
	ctx := context.Background()

	rl := role.New(
		uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		name.MustParse("Admin"),
		time.Time{},
		time.Time{},
		nil,
	)

	mLogger := logger.NewMockLogger(t)

	mStorer := role.NewMockStorer(t)
	mStorer.EXPECT().Delete(ctx, rl).Return(errors.New("delete error"))

	sut := role_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	err := sut.Delete(ctx, rl)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delete error")
}

func TestService_QueryByID_Successful(t *testing.T) {
	ctx := context.Background()
	roleID := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	expectedRole := role.New(
		roleID,
		name.MustParse("Admin"),
		mTime,
		mTime,
		nil,
	)

	mLogger := logger.NewMockLogger(t)

	mStorer := role.NewMockStorer(t)
	mStorer.EXPECT().QueryByID(ctx, roleID).Return(expectedRole, nil)

	sut := role_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	rl, err := sut.QueryByID(ctx, roleID)

	assert.NoError(t, err)
	assert.Equal(t, expectedRole.ID(), rl.ID())
	assert.Equal(t, expectedRole.Name(), rl.Name())
}

func TestService_QueryByID_NotFound(t *testing.T) {
	ctx := context.Background()
	roleID := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")

	mLogger := logger.NewMockLogger(t)

	mStorer := role.NewMockStorer(t)
	mStorer.EXPECT().QueryByID(ctx, roleID).Return(role.Role{}, role.ErrNotFound)

	sut := role_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	_, err := sut.QueryByID(ctx, roleID)

	assert.Error(t, err)
	assert.ErrorIs(t, err, role.ErrNotFound)
}

func TestService_Query_Successful(t *testing.T) {
	ctx := context.Background()
	filter := role.QueryFilter{}
	orderBy := order.By{Field: "name", Direction: "asc"}
	cur := cursor.Cursor{}

	expectedRoles := []role.Role{
		role.New(
			uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
			name.MustParse("Admin"),
			time.Time{},
			time.Time{},
			nil,
		),
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := role.NewMockStorer(t)
	mStorer.EXPECT().Query(ctx, filter, orderBy, cur).Return(expectedRoles, nil)

	sut := role_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	roles, err := sut.Query(ctx, filter, orderBy, cur)

	assert.NoError(t, err)
	assert.Len(t, roles, 1)
	assert.Equal(t, expectedRoles[0].Name(), roles[0].Name())
}

func TestService_Query_Error(t *testing.T) {
	ctx := context.Background()
	filter := role.QueryFilter{}
	orderBy := order.By{Field: "name", Direction: "asc"}
	cur := cursor.Cursor{}

	mLogger := logger.NewMockLogger(t)

	mStorer := role.NewMockStorer(t)
	mStorer.EXPECT().Query(ctx, filter, orderBy, cur).Return(nil, errors.New("query error"))

	sut := role_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	_, err := sut.Query(ctx, filter, orderBy, cur)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "query error")
}
