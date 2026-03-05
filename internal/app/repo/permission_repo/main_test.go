package permission_repo_test

import (
	"context"
	"os"
	"testing"

	"github.com/Housiadas/cerberus/internal/utils/dbtest"
)

var sc *dbtest.SharedContainer

func TestMain(m *testing.M) {
	var teardown func()
	sc, teardown = dbtest.SetupSharedContainer(context.Background())
	defer teardown()
	os.Exit(m.Run())
}
