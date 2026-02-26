package permission_service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Housiadas/cerberus/internal/utils/page"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Housiadas/cerberus/internal/core/domain/name"
	"github.com/Housiadas/cerberus/internal/core/domain/permission"
	"github.com/Housiadas/cerberus/internal/core/service/permission_service"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
)

func TestService_Create_Successful(t *testing.T) {
	ctx := context.Background()

	np := permission.NewPermission{
		Name: name.MustParse("Read Users"),
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := permission.NewMockStorer(t)
	mStorer.EXPECT().Create(ctx, mock.AnythingOfType("permission.Permission")).Return(nil)

	sut := permission_service.New(mLogger, mStorer)
	p, err := sut.Create(ctx, np)

	assert.NoError(t, err)
	assert.Equal(t, np.Name, p.Name)
	assert.NotZero(t, p.CreatedAt)
	assert.NotZero(t, p.UpdatedAt)
}

func TestService_Create_StorerError(t *testing.T) {
	ctx := context.Background()

	np := permission.NewPermission{
		Name: name.MustParse("Read Users"),
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := permission.NewMockStorer(t)
	mStorer.EXPECT().Create(ctx, mock.AnythingOfType("permission.Permission")).Return(errors.New("create error"))

	sut := permission_service.New(mLogger, mStorer)
	_, err := sut.Create(ctx, np)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "create error")
}

func TestService_Update_Successful(t *testing.T) {
	ctx := context.Background()
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	existingPerm := permission.Permission{
		ID:        uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		Name:      name.MustParse("Read Users"),
		CreatedAt: mTime,
		UpdatedAt: mTime,
	}

	newName := name.MustParse("Write Users")
	up := permission.UpdatePermission{
		Name: &newName,
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := permission.NewMockStorer(t)
	mStorer.EXPECT().Update(ctx, mock.AnythingOfType("permission.Permission")).Return(nil)

	sut := permission_service.New(mLogger, mStorer)
	p, err := sut.Update(ctx, existingPerm, up)

	assert.NoError(t, err)
	assert.Equal(t, newName, p.Name)
}

func TestService_Update_StorerError(t *testing.T) {
	ctx := context.Background()
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	existingPerm := permission.Permission{
		ID:        uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		Name:      name.MustParse("Read Users"),
		CreatedAt: mTime,
		UpdatedAt: mTime,
	}

	newName := name.MustParse("Write Users")
	up := permission.UpdatePermission{
		Name: &newName,
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := permission.NewMockStorer(t)
	mStorer.EXPECT().Update(ctx, mock.AnythingOfType("permission.Permission")).Return(errors.New("update error"))

	sut := permission_service.New(mLogger, mStorer)
	_, err := sut.Update(ctx, existingPerm, up)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update error")
}

func TestService_Delete_Successful(t *testing.T) {
	ctx := context.Background()

	p := permission.Permission{
		ID:   uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		Name: name.MustParse("Read Users"),
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := permission.NewMockStorer(t)
	mStorer.EXPECT().Delete(ctx, p).Return(nil)

	sut := permission_service.New(mLogger, mStorer)
	err := sut.Delete(ctx, p)

	assert.NoError(t, err)
}

func TestService_Delete_Error(t *testing.T) {
	ctx := context.Background()

	p := permission.Permission{
		ID:   uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
		Name: name.MustParse("Read Users"),
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := permission.NewMockStorer(t)
	mStorer.EXPECT().Delete(ctx, p).Return(errors.New("delete error"))

	sut := permission_service.New(mLogger, mStorer)
	err := sut.Delete(ctx, p)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delete error")
}

func TestService_Count_Successful(t *testing.T) {
	ctx := context.Background()
	filter := permission.QueryFilter{}

	mLogger := logger.NewMockLogger(t)

	mStorer := permission.NewMockStorer(t)
	mStorer.EXPECT().Count(ctx, filter).Return(5, nil)

	sut := permission_service.New(mLogger, mStorer)
	count, err := sut.Count(ctx, filter)

	assert.NoError(t, err)
	assert.Equal(t, 5, count)
}

func TestService_Count_Error(t *testing.T) {
	ctx := context.Background()
	filter := permission.QueryFilter{}

	mLogger := logger.NewMockLogger(t)

	mStorer := permission.NewMockStorer(t)
	mStorer.EXPECT().Count(ctx, filter).Return(0, errors.New("count error"))

	sut := permission_service.New(mLogger, mStorer)
	_, err := sut.Count(ctx, filter)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "count error")
}

func TestService_QueryByID_Successful(t *testing.T) {
	ctx := context.Background()
	permID := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")
	mTime := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)

	expectedPerm := permission.Permission{
		ID:        permID,
		Name:      name.MustParse("Read Users"),
		CreatedAt: mTime,
		UpdatedAt: mTime,
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := permission.NewMockStorer(t)
	mStorer.EXPECT().QueryByID(ctx, permID).Return(expectedPerm, nil)

	sut := permission_service.New(mLogger, mStorer)
	p, err := sut.QueryByID(ctx, permID)

	assert.NoError(t, err)
	assert.Equal(t, expectedPerm.ID, p.ID)
	assert.Equal(t, expectedPerm.Name, p.Name)
}

func TestService_QueryByID_NotFound(t *testing.T) {
	ctx := context.Background()
	permID := uuid.MustParse("01234567-89ab-7def-0123-456789abcdef")

	mLogger := logger.NewMockLogger(t)

	mStorer := permission.NewMockStorer(t)
	mStorer.EXPECT().QueryByID(ctx, permID).Return(permission.Permission{}, permission.ErrNotFound)

	sut := permission_service.New(mLogger, mStorer)
	_, err := sut.QueryByID(ctx, permID)

	assert.Error(t, err)
	assert.ErrorIs(t, err, permission.ErrNotFound)
}

func TestService_Query_Successful(t *testing.T) {
	ctx := context.Background()
	filter := permission.QueryFilter{}
	orderBy := order.By{Field: "name", Direction: "asc"}
	page := page.Page{}

	expectedPerms := []permission.Permission{
		{
			ID:   uuid.MustParse("01234567-89ab-7def-0123-456789abcdef"),
			Name: name.MustParse("Read Users"),
		},
	}

	mLogger := logger.NewMockLogger(t)

	mStorer := permission.NewMockStorer(t)
	mStorer.EXPECT().Query(ctx, filter, orderBy, page).Return(expectedPerms, nil)

	sut := permission_service.New(mLogger, mStorer)
	perms, err := sut.Query(ctx, filter, orderBy, page)

	assert.NoError(t, err)
	assert.Len(t, perms, 1)
	assert.Equal(t, expectedPerms[0].Name, perms[0].Name)
}

func TestService_Query_Error(t *testing.T) {
	ctx := context.Background()
	filter := permission.QueryFilter{}
	orderBy := order.By{Field: "name", Direction: "asc"}
	page := page.Page{}

	mLogger := logger.NewMockLogger(t)

	mStorer := permission.NewMockStorer(t)
	mStorer.EXPECT().Query(ctx, filter, orderBy, page).Return(nil, errors.New("query error"))

	sut := permission_service.New(mLogger, mStorer)
	_, err := sut.Query(ctx, filter, orderBy, page)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "query error")
}
