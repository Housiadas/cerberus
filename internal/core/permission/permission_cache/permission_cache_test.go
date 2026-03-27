package permission_cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/Housiadas/cerberus/internal/core/permission"
	"github.com/Housiadas/cerberus/internal/core/permission/permission_cache"
	"github.com/Housiadas/cerberus/internal/types/name"
	goredis "github.com/redis/go-redis/v9"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newTestStore(t *testing.T) (*permission_cache.Store, *mockstorer, *mockredisService) {
	t.Helper()

	mockStorer := newMockstorer(t)
	mLogger := newMocklogger(t)
	mRed := newMockredisService(t)
	store := permission_cache.NewStore(t.Context(), mLogger, mockStorer, mRed)

	return store, mockStorer, mRed
}

func redisCacheMiss(red *redis.MockClient) {
	red.On("Get", mock.Anything, mock.Anything).
		Return(goredis.NewStringResult("", goredis.Nil)).Once()
	red.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(goredis.NewStatusResult("OK", nil)).Maybe()
}

func TestQueryByID_CacheMiss(t *testing.T) {
	store, mockStorer, red := newTestStore(t)
	ctx := context.Background()

	id := uuid.New()
	expected := permission.New(id, name.Name{}, time.Time{}, time.Time{}, nil)

	redisCacheMiss(red)
	mockStorer.On("QueryByID", ctx, id).Return(expected, nil)

	result, err := store.QueryByID(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestQueryByID_CacheHit(t *testing.T) {
	store, mockStorer, red := newTestStore(t)
	ctx := context.Background()

	id := uuid.New()
	expected := permission.New(id, name.Name{}, time.Time{}, time.Time{}, nil)

	redisCacheMiss(red)
	mockStorer.On("QueryByID", ctx, id).Return(expected, nil).Once()

	result, err := store.QueryByID(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, expected, result)

	result, err = store.QueryByID(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestCreate_DelegatesToStorer(t *testing.T) {
	store, mockStorer, _ := newTestStore(t)
	ctx := context.Background()

	p := permission.New(uuid.New(), name.Name{}, time.Time{}, time.Time{}, nil)

	mockStorer.On("Create", ctx, p).Return(nil)

	err := store.Create(ctx, p)
	require.NoError(t, err)
}

func TestUpdate_DelegatesToStorerAndInvalidatesCache(t *testing.T) {
	store, mockStorer, _ := newTestStore(t)
	ctx := context.Background()

	p := permission.New(uuid.New(), name.Name{}, time.Time{}, time.Time{}, nil)

	mockStorer.On("Update", ctx, p).Return(nil)

	err := store.Update(ctx, p)
	require.NoError(t, err)
}

func TestDelete_DelegatesToStorerAndInvalidatesCache(t *testing.T) {
	store, mockStorer, _ := newTestStore(t)
	ctx := context.Background()

	p := permission.New(uuid.New(), name.Name{}, time.Time{}, time.Time{}, nil)

	mockStorer.On("Delete", ctx, p).Return(nil)

	err := store.Delete(ctx, p)
	require.NoError(t, err)
}
