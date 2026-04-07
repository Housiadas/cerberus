package commands_test

import (
	"context"
	"os"
	"testing"

	"github.com/Housiadas/cerberus/internal/sdk/testutil/dbtest"
	"github.com/Housiadas/cerberus/internal/sdk/testutil/kafkatest"
	ckafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

var (
	sc          *dbtest.SharedContainer
	sharedKafka producer
)

type producer interface {
	Produce(ctx context.Context, msg *ckafka.Message) error
	Flush(timeoutMs int) int
	Close()
}

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
