package redis

import (
	"context"
	"testing"
	"time"

	goRedis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestDistributedStorage_Get_Hit(t *testing.T) {
	client := NewMockClient(t)
	ds := NewDistributedStorage(client, 5*time.Minute)

	ctx := context.Background()

	cmd := goRedis.NewStringResult("test-value", nil)
	client.On("Get", ctx, "key1").Return(cmd)

	val, ok := ds.Get(ctx, "key1")
	assert.True(t, ok)
	assert.Equal(t, []byte("test-value"), val)
}

func TestDistributedStorage_Get_Miss(t *testing.T) {
	client := NewMockClient(t)
	ds := NewDistributedStorage(client, 5*time.Minute)

	ctx := context.Background()

	cmd := goRedis.NewStringResult("", goRedis.Nil)
	client.On("Get", ctx, "key1").Return(cmd)

	val, ok := ds.Get(ctx, "key1")
	assert.False(t, ok)
	assert.Nil(t, val)
}

func TestDistributedStorage_Set(t *testing.T) {
	client := NewMockClient(t)
	ds := NewDistributedStorage(client, 5*time.Minute)

	ctx := context.Background()

	cmd := goRedis.NewStatusResult("OK", nil)
	client.On("Set", ctx, "key1", []byte("value1"), 5*time.Minute).Return(cmd)

	ds.Set(ctx, "key1", []byte("value1"))
}

func TestDistributedStorage_GetBatch_Partial(t *testing.T) {
	client := NewMockClient(t)
	ds := NewDistributedStorage(client, 5*time.Minute)

	ctx := context.Background()

	cmd := goRedis.NewSliceResult([]any{"val1", nil, "val3"}, nil)
	client.On("MGet", ctx, []string{"k1", "k2", "k3"}).Return(cmd)

	result := ds.GetBatch(ctx, []string{"k1", "k2", "k3"})
	assert.Len(t, result, 2)
	assert.Equal(t, []byte("val1"), result["k1"])
	assert.Equal(t, []byte("val3"), result["k3"])
}

func TestDistributedStorage_GetBatch_Empty(t *testing.T) {
	client := NewMockClient(t)
	ds := NewDistributedStorage(client, 5*time.Minute)

	ctx := context.Background()

	result := ds.GetBatch(ctx, []string{})
	assert.Nil(t, result)
}

func TestDistributedStorage_SetBatch_Empty(t *testing.T) {
	client := NewMockClient(t)
	ds := NewDistributedStorage(client, 5*time.Minute)

	ctx := context.Background()

	ds.SetBatch(ctx, map[string][]byte{})
}
