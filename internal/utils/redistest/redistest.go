// Package redistest provides support for running tests that use Redis.
package redistest

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Housiadas/cerberus/internal/config"
	pkgRedis "github.com/Housiadas/cerberus/pkg/redis"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcRedis "github.com/testcontainers/testcontainers-go/modules/redis"
)

// New starts a Redis container using testcontainers and returns
// a configured DistributedStorage connected to it.
func New(t *testing.T) *pkgRedis.DistributedStorage {
	t.Helper()

	ctx := context.Background()

	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	ctr, err := tcRedis.Run(ctx, cfg.Redis.RedisImage)
	defer testcontainers.CleanupContainer(t, ctr)

	require.NoError(t, err)

	connStr, err := ctr.ConnectionString(ctx)
	require.NoError(t, err)

	client, err := pkgRedis.Open(ctx, pkgRedis.Config{
		Host: strings.TrimPrefix(connStr, "redis://"),
	})
	require.NoError(t, err)

	ttl, err := time.ParseDuration(cfg.Redis.TTL)
	require.NoError(t, err)

	t.Cleanup(func() {
		client.Close()
	})

	return pkgRedis.NewDistributedStorage(client, ttl)
}
