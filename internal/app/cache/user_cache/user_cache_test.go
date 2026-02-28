package user_cache

import (
	"context"
	"net/mail"
	"testing"

	"github.com/Housiadas/cerberus/internal/core/domain/user"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/redis"
	goredis "github.com/redis/go-redis/v9"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newTestStore(t *testing.T) (*Store, *user.MockStorer, *redis.MockClient) {
	t.Helper()

	mockStorer := user.NewMockStorer(t)
	log := logger.NewMockLogger(t)
	red := redis.NewMockClient(t)
	store := NewStore(log, mockStorer, red)

	return store, mockStorer, red
}

// redisCacheMiss sets up the redis mock to simulate an L2 cache miss (Get returns redis.Nil)
// and then a Set to store the fetched value.
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
	expected := user.User{
		ID:    id,
		Email: mail.Address{Address: "test@example.com"},
	}

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
	expected := user.User{
		ID:    id,
		Email: mail.Address{Address: "test@example.com"},
	}

	redisCacheMiss(red)
	mockStorer.On("QueryByID", ctx, id).Return(expected, nil).Once()

	// First call - cache miss, fetches from storer
	result, err := store.QueryByID(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, expected, result)

	// Second call - should be served from L1 cache (neither storer nor redis called again)
	result, err = store.QueryByID(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestQueryByEmail_CacheMiss(t *testing.T) {
	store, mockStorer, red := newTestStore(t)
	ctx := context.Background()

	email := mail.Address{Address: "test@example.com"}
	expected := user.User{
		ID:    uuid.New(),
		Email: email,
	}

	redisCacheMiss(red)
	mockStorer.On("QueryByEmail", ctx, email).Return(expected, nil)

	result, err := store.QueryByEmail(ctx, email)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestCreate_DelegatesToStorer(t *testing.T) {
	store, mockStorer, _ := newTestStore(t)
	ctx := context.Background()

	usr := user.User{
		ID:    uuid.New(),
		Email: mail.Address{Address: "test@example.com"},
	}

	mockStorer.On("Create", ctx, usr).Return(nil)

	err := store.Create(ctx, usr)
	require.NoError(t, err)
}

func TestUpdate_DelegatesToStorerAndInvalidatesCache(t *testing.T) {
	store, mockStorer, _ := newTestStore(t)
	ctx := context.Background()

	usr := user.User{
		ID:    uuid.New(),
		Email: mail.Address{Address: "test@example.com"},
	}

	mockStorer.On("Update", ctx, usr).Return(nil)

	err := store.Update(ctx, usr)
	require.NoError(t, err)
}

func TestDelete_DelegatesToStorerAndInvalidatesCache(t *testing.T) {
	store, mockStorer, _ := newTestStore(t)
	ctx := context.Background()

	usr := user.User{
		ID:    uuid.New(),
		Email: mail.Address{Address: "test@example.com"},
	}

	mockStorer.On("Delete", ctx, usr).Return(nil)

	err := store.Delete(ctx, usr)
	require.NoError(t, err)
}
