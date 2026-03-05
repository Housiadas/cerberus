package outbox_relay_test

import (
	"context"
	"os"
	"testing"

	"github.com/Housiadas/cerberus/internal/utils/dbtest"
	"github.com/Housiadas/cerberus/internal/utils/kafkatest"
	"github.com/Housiadas/cerberus/pkg/kafka"
)

var (
	sc          *dbtest.SharedContainer
	sharedKafka kafka.Producer
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	var pgTeardown func()
	sc, pgTeardown = dbtest.SetupSharedContainer(ctx)

	var kafkaTeardown func()
	sharedKafka, kafkaTeardown = kafkatest.SetupSharedContainer(ctx)

	defer pgTeardown()
	defer kafkaTeardown()

	os.Exit(m.Run())
}
