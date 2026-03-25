package permission_service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	permission2 "github.com/Housiadas/cerberus/internal/core/permission"
	"github.com/Housiadas/cerberus/internal/core/permission/permission_service"
	"github.com/Housiadas/cerberus/internal/types/name"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
)

func TestService_Create_Successful(t *testing.T) {
	ctx := context.Background()

	np := permission2.NewPermission{
		Name: name.MustParse("Read Users"),
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := permission2.NewMockStorer(t)
	mStorer.EXPECT().Create(ctx, mock.AnythingOfType("permission.Permission")).Return(nil)

	mUUID := uuidgen.NewMockGenerator(t)
	mUUID.EXPECT().Generate().Return(uuid.New(), nil)

	sut := permission_service.New(mLogger, mStorer, mUUID)
	p, err := sut.Create(ctx, np)

	assert.NoError(t, err)
	assert.Equal(t, np.Name, p.Name())
	assert.NotZero(t, p.CreatedAt())
	assert.NotZero(t, p.UpdatedAt())
}

func TestService_Create_StorerError(t *testing.T) {
	ctx := context.Background()

	np := permission2.NewPermission{
		Name: name.MustParse("Read Users"),
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := permission2.NewMockStorer(t)
	mStorer.EXPECT().Create(ctx, mock.AnythingOfType("permission.Permission")).Return(errors.New("create error"))

	mUUID := uuidgen.NewMockGenerator(t)
	mUUID.EXPECT().Generate().Return(uuid.New(), nil)

	sut := permission_service.New(mLogger, mStorer, mUUID)
	_, err := sut.Create(ctx, np)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "create error")
}

func TestService_Update_Successful(t *testing.T) {
	ctx := context.Background()
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	existingPerm := permission2.New(
		uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		name.MustParse("Read Users"),
		mTime,
		mTime,
		nil,
	)

	newName := name.MustParse("Write Users")
	up := permission2.UpdatePermission{
		Name: &newName,
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := permission2.NewMockStorer(t)
	mStorer.EXPECT().Update(ctx, mock.AnythingOfType("permission.Permission")).Return(nil)

	sut := permission_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	p, err := sut.Update(ctx, existingPerm, up)

	assert.NoError(t, err)
	assert.Equal(t, newName, p.Name())
}

func TestService_Update_StorerError(t *testing.T) {
	ctx := context.Background()
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	existingPerm := permission2.New(
		uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		name.MustParse("Read Users"),
		mTime,
		mTime,
		nil,
	)

	newName := name.MustParse("Write Users")
	up := permission2.UpdatePermission{
		Name: &newName,
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := permission2.NewMockStorer(t)
	mStorer.EXPECT().Update(ctx, mock.AnythingOfType("permission.Permission")).Return(errors.New("update error"))

	sut := permission_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	_, err := sut.Update(ctx, existingPerm, up)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update error")
}

func TestService_Delete_Successful(t *testing.T) {
	ctx := context.Background()

	p := permission2.New(
		uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		name.MustParse("Read Users"),
		time.Time{},
		time.Time{},
		nil,
	)

	mLogger := logger.NewMockLogger(t)

	mStorer := permission2.NewMockStorer(t)
	mStorer.EXPECT().Delete(ctx, p).Return(nil)

	sut := permission_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	err := sut.Delete(ctx, p)

	assert.NoError(t, err)
}

func TestService_Delete_Error(t *testing.T) {
	ctx := context.Background()

	p := permission2.New(
		uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		name.MustParse("Read Users"),
		time.Time{},
		time.Time{},
		nil,
	)

	mLogger := logger.NewMockLogger(t)

	mStorer := permission2.NewMockStorer(t)
	mStorer.EXPECT().Delete(ctx, p).Return(errors.New("delete error"))

	sut := permission_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	err := sut.Delete(ctx, p)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delete error")
}

func TestService_QueryByID_Successful(t *testing.T) {
	ctx := context.Background()
	permID := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	expectedPerm := permission2.New(
		permID,
		name.MustParse("Read Users"),
		mTime,
		mTime,
		nil,
	)

	mLogger := logger.NewMockLogger(t)

	mStorer := permission2.NewMockStorer(t)
	mStorer.EXPECT().QueryByID(ctx, permID).Return(expectedPerm, nil)

	sut := permission_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	p, err := sut.QueryByID(ctx, permID)

	assert.NoError(t, err)
	assert.Equal(t, expectedPerm.ID(), p.ID())
	assert.Equal(t, expectedPerm.Name(), p.Name())
}

func TestService_QueryByID_NotFound(t *testing.T) {
	ctx := context.Background()
	permID := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")

	mLogger := logger.NewMockLogger(t)

	mStorer := permission2.NewMockStorer(t)
	mStorer.EXPECT().QueryByID(ctx, permID).Return(permission2.Permission{}, permission2.ErrNotFound)

	sut := permission_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	_, err := sut.QueryByID(ctx, permID)

	assert.Error(t, err)
	assert.ErrorIs(t, err, permission2.ErrNotFound)
}

func TestService_Query_Successful(t *testing.T) {
	ctx := context.Background()
	filter := permission2.QueryFilter{}
	orderBy := order.By{Field: "name", Direction: "asc"}
	cur := cursor.Cursor{}

	expectedPerms := []permission2.Permission{
		permission2.New(
			uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
			name.MustParse("Read Users"),
			time.Time{},
			time.Time{},
			nil,
		),
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := permission2.NewMockStorer(t)
	mStorer.EXPECT().Query(ctx, filter, orderBy, cur).Return(expectedPerms, nil)

	sut := permission_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	perms, err := sut.Query(ctx, filter, orderBy, cur)

	assert.NoError(t, err)
	assert.Len(t, perms, 1)
	assert.Equal(t, expectedPerms[0].Name(), perms[0].Name())
}

func TestService_Query_Error(t *testing.T) {
	ctx := context.Background()
	filter := permission2.QueryFilter{}
	orderBy := order.By{Field: "name", Direction: "asc"}
	cur := cursor.Cursor{}

	mLogger := logger.NewMockLogger(t)

	mStorer := permission2.NewMockStorer(t)
	mStorer.EXPECT().Query(ctx, filter, orderBy, cur).Return(nil, errors.New("query error"))

	sut := permission_service.New(mLogger, mStorer, uuidgen.NewMockGenerator(t))
	_, err := sut.Query(ctx, filter, orderBy, cur)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "query error")
}
