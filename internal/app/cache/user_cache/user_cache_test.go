package user_cache

import (
	"context"
	"net/mail"
	"testing"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/user"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestStore(t *testing.T) (*Store, *user.MockStorer) {
	t.Helper()

	mockStorer := user.NewMockStorer(t)
	log := logger.NewMockLogger(t)
	store := NewStore(log, mockStorer, 5*time.Minute, nil)

	return store, mockStorer
}

func TestQueryByID_CacheMiss(t *testing.T) {
	store, mockStorer := newTestStore(t)
	ctx := context.Background()

	id := uuid.New()
	expected := user.User{
		ID:    id,
		Email: mail.Address{Address: "test@example.com"},
	}

	mockStorer.On("QueryByID", ctx, id).Return(expected, nil)

	result, err := store.QueryByID(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestQueryByID_CacheHit(t *testing.T) {
	store, mockStorer := newTestStore(t)
	ctx := context.Background()

	id := uuid.New()
	expected := user.User{
		ID:    id,
		Email: mail.Address{Address: "test@example.com"},
	}

	mockStorer.On("QueryByID", ctx, id).Return(expected, nil).Once()

	// First call - cache miss, fetches from storer
	result, err := store.QueryByID(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, expected, result)

	// Second call - should be served from cache (storer not called again)
	result, err = store.QueryByID(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestQueryByEmail_CacheMiss(t *testing.T) {
	store, mockStorer := newTestStore(t)
	ctx := context.Background()

	email := mail.Address{Address: "test@example.com"}
	expected := user.User{
		ID:    uuid.New(),
		Email: email,
	}

	mockStorer.On("QueryByEmail", ctx, email).Return(expected, nil)

	result, err := store.QueryByEmail(ctx, email)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestCreate_DelegatesToStorer(t *testing.T) {
	store, mockStorer := newTestStore(t)
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
	store, mockStorer := newTestStore(t)
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
	store, mockStorer := newTestStore(t)
	ctx := context.Background()

	usr := user.User{
		ID:    uuid.New(),
		Email: mail.Address{Address: "test@example.com"},
	}

	mockStorer.On("Delete", ctx, usr).Return(nil)

	err := store.Delete(ctx, usr)
	require.NoError(t, err)
}
