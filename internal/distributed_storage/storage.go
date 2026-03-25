package distributed_storage

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client defines the interface for Redis operations used by distributed storage.
type client interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd
	MGet(ctx context.Context, keys ...string) *redis.SliceCmd
	Pipeline() redis.Pipeliner
}

// DistributedStorage implements sturdyc.DistributedStorage using Redis.
type DistributedStorage struct {
	client client
	ttl    time.Duration
}

// New creates a new DistributedStorage backed by Redis.
func New(
	client client,
	ttl time.Duration,
) *DistributedStorage {
	return &DistributedStorage{
		client: client,
		ttl:    ttl,
	}
}

// Get retrieves a value by key from Redis. Returns false on cache miss.
func (ds *DistributedStorage) Get(ctx context.Context, key string) ([]byte, bool) {
	val, err := ds.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, false
		}

		return nil, false
	}

	return val, true
}

// Set stores a key-value pair in Redis with the configured TTL.
func (ds *DistributedStorage) Set(ctx context.Context, key string, value []byte) {
	ds.client.Set(ctx, key, value, ds.ttl)
}

// GetBatch retrieves multiple values by keys from Redis.
func (ds *DistributedStorage) GetBatch(ctx context.Context, keys []string) map[string][]byte {
	if len(keys) == 0 {
		return nil
	}

	vals, err := ds.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil
	}

	result := make(map[string][]byte)

	for i, val := range vals {
		if val == nil {
			continue
		}

		str, ok := val.(string)
		if !ok {
			continue
		}

		result[keys[i]] = []byte(str)
	}

	return result
}

// SetBatch stores multiple key-value pairs in Redis using a pipeline.
func (ds *DistributedStorage) SetBatch(ctx context.Context, records map[string][]byte) {
	if len(records) == 0 {
		return
	}

	pipe := ds.client.Pipeline()

	for key, value := range records {
		pipe.Set(ctx, key, value, ds.ttl)
	}

	_, _ = pipe.Exec(ctx)
}
