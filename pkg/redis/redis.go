// Package redis provides support for accessing a Redis database.
package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Config is the required properties to use Redis.
type Config struct {
	Host     string
	Password string
	DB       int
}

// Client defines the interface for Redis operations used by distributed storage.
type Client interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd
	MGet(ctx context.Context, keys ...string) *redis.SliceCmd
	Pipeline() redis.Pipeliner
}

// Open knows how to open a Redis connection based on the configuration.
func Open(ctx context.Context, cfg Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Host,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := client.Ping(pingCtx).Err()
	if err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	return client, nil
}
