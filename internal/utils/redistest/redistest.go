// Package redistest provides support for running tests that use Redis.
package redistest

import (
	"context"
	"strings"

	"github.com/Housiadas/cerberus/internal/config"
	pkgRedis "github.com/Housiadas/cerberus/pkg/redis"
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
) pkgRedis.Client {
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
		err2 := testcontainers.TerminateContainer(ctr)
		if err2 != nil {
			t.Logf("terminate shared redis container: %s", err2)
		}
	})

	connStr, err := ctr.ConnectionString(ctx)
	if err != nil {
		t.Fatal("shared redis connection string:", err)
	}

	client, err := pkgRedis.Open(ctx, pkgRedis.Config{
		Host: strings.TrimPrefix(connStr, "redis://"),
	})
	if err != nil {
		t.Fatal("open shared redis client:", err)
	}

	t.Cleanup(func() {
		client.Close()
	})

	return client
}
