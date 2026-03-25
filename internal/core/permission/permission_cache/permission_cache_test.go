package permission_cache

import (
	"context"
	"testing"
	"time"

	permission2 "github.com/Housiadas/cerberus/internal/core/permission"
	"github.com/Housiadas/cerberus/internal/types/name"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/redis"
	goredis "github.com/redis/go-redis/v9"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newTestStore(t *testing.T) (*Store, *permission2.MockStorer, *redis.MockClient) {
	t.Helper()

	mockStorer := permission2.NewMockStorer(t)
	mLogger := logger.NewMockLogger(t)
	mRed := redis.NewMockClient(t)
	store := NewStore(t.Context(), mLogger, mockStorer, mRed)

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
	expected := permission2.New(id, name.Name{}, time.Time{}, time.Time{}, nil)

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
	expected := permission2.New(id, name.Name{}, time.Time{}, time.Time{}, nil)

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

	p := permission2.New(uuid.New(), name.Name{}, time.Time{}, time.Time{}, nil)

	mockStorer.On("Create", ctx, p).Return(nil)

	err := store.Create(ctx, p)
	require.NoError(t, err)
}

func TestUpdate_DelegatesToStorerAndInvalidatesCache(t *testing.T) {
	store, mockStorer, _ := newTestStore(t)
	ctx := context.Background()

	p := permission2.New(uuid.New(), name.Name{}, time.Time{}, time.Time{}, nil)

	mockStorer.On("Update", ctx, p).Return(nil)

	err := store.Update(ctx, p)
	require.NoError(t, err)
}

func TestDelete_DelegatesToStorerAndInvalidatesCache(t *testing.T) {
	store, mockStorer, _ := newTestStore(t)
	ctx := context.Background()

	p := permission2.New(uuid.New(), name.Name{}, time.Time{}, time.Time{}, nil)

	mockStorer.On("Delete", ctx, p).Return(nil)

	err := store.Delete(ctx, p)
	require.NoError(t, err)
}
