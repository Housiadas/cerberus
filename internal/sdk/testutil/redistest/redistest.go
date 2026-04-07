// Package redistest provides support for running tests that use Redis.
package redistest

import (
	"context"
	"strings"
	"time"

	"github.com/Housiadas/cerberus/internal/config"
	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	tcRedis "github.com/testcontainers/testcontainers-go/modules/redis"
)

// NewSharedContainer starts a single Redis container intended to be shared
// across all tests in a package via TestMain.
func NewSharedContainer(ctx context.Context, t interface {
	Helper()
	Fatal(args ...any)
	Logf(format string, args ...any)
	Cleanup(fn func())
},
) *redis.Client {
	t.Helper()

	appCfg, err := config.LoadConfig()
	if err != nil {
		t.Fatal("load config:", err)
	}

	ctr, err := tcRedis.Run(ctx, appCfg.Redis.RedisImage)
	if err != nil {
		t.Fatal("start shared redis container:", err)
	}

	t.Cleanup(func() {
		err := testcontainers.TerminateContainer(ctr)
		if err != nil {
			t.Logf("terminate shared redis container: %s", err)
		}
	})

	connStr, err := ctr.ConnectionString(ctx)
	if err != nil {
		t.Fatal("shared redis connection string:", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr: strings.TrimPrefix(connStr, "redis://"),
	})

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = client.Ping(pingCtx).Err()
	if err != nil {
		t.Fatal("open redis client:", err)
	}

	t.Cleanup(func() {
		client.Close()
	})

	return client
}
