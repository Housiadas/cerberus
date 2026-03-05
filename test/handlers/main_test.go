package handlers_test

import (
	"os"
	"testing"

	"github.com/Housiadas/cerberus/internal/utils/apitest"
)

var env *apitest.Env

func TestMain(m *testing.M) {
	var teardown func()
	env, teardown = apitest.SetupEnv()
	defer teardown()
	os.Exit(m.Run())
}
