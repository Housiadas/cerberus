// Package kafkatest provides support for running tests that use Kafka.
package kafkatest

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/pkg/kafka"
	ckafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/testcontainers/testcontainers-go"
	tcKafka "github.com/testcontainers/testcontainers-go/modules/kafka"
)

const (
	KafkaImage     = "confluentinc/confluent-local:7.5.0"
	KafkaClusterID = "test-cluster"
)

type producer interface {
	Produce(ctx context.Context, msg *ckafka.Message) error
	Flush(timeoutMs int) int
	Close()
}

// SetupSharedContainer starts a Kafka container for use in TestMain.
// Returns the Producer and a teardown function to call via defer.
func SetupSharedContainer(ctx context.Context) (*kafka.ProducerClient, func()) {
	t := &mainT{}
	producer := NewSharedContainer(ctx, t)
	teardown := func() {
		for i := len(t.cleanups) - 1; i >= 0; i-- {
			t.cleanups[i]()
		}
	}

	return producer, teardown
}

// mainT is a minimal testing-like facade usable from TestMain.
type mainT struct {
	cleanups []func()
}

func (m *mainT) Helper() {}

func (m *mainT) Fatal(args ...any) {
	panic(fmt.Sprint(args...))
}

func (m *mainT) Logf(format string, args ...any) {
	fmt.Printf(format+"\n", args...)
}

func (m *mainT) Cleanup(fn func()) {
	m.cleanups = append(m.cleanups, fn)
}

// NewSharedContainer starts a single Kafka container intended to be shared
// across all tests in a package via TestMain.
func NewSharedContainer(ctx context.Context, t interface {
	Helper()
	Fatal(args ...any)
	Logf(format string, args ...any)
	Cleanup(fn func())
},
) *kafka.ProducerClient {
	t.Helper()

	cfg := newConfig(t)

	ctr, err := tcKafka.Run(
		ctx,
		KafkaImage,
		tcKafka.WithClusterID(KafkaClusterID),
	)
	if err != nil {
		t.Fatal("start shared kafka container:", err)
	}

	t.Cleanup(func() {
		err2 := testcontainers.TerminateContainer(ctr)
		if err2 != nil {
			t.Logf("terminate shared kafka container: %s", err2)
		}
	})

	brokers, err := ctr.Brokers(ctx)
	if err != nil {
		t.Fatal("shared kafka brokers:", err)
	}

	prod, err := kafka.NewProducer(kafka.ProducerConfig{
		Brokers:          brokers[0],
		AddressFamily:    cfg.AddressFamily,
		SecurityProtocol: cfg.SecurityProtocol,
		LogLevel:         cfg.LogLevel,
		MaxMessageBytes:  cfg.MaxMessageBytes,
	})
	if err != nil {
		t.Fatal("create shared kafka producer:", err)
	}

	t.Cleanup(func() {
		prod.Close()
	})

	return prod
}
