// Package kafkatest provides support for running tests that use Kafka.
package kafkatest

import (
	"context"
	"testing"

	"github.com/Housiadas/cerberus/pkg/kafka"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcKafka "github.com/testcontainers/testcontainers-go/modules/kafka"
)

const (
	KafkaImage     = "confluentinc/confluent-local:7.5.0"
	KafkaClusterID = "test-cluster"
)

// New starts a Kafka container using testcontainers and returns
// a configured kafka.Producer connected to it.
func New(t *testing.T) kafka.Producer {
	t.Helper()

	ctx := context.Background()

	// load app local config
	cfg := newConfig(t)

	ctr, err := tcKafka.Run(
		ctx,
		KafkaImage,
		tcKafka.WithClusterID(KafkaClusterID),
	)
	defer testcontainers.CleanupContainer(t, ctr)

	require.NoError(t, err)

	brokers, err := ctr.Brokers(ctx)
	require.NoError(t, err)

	producer, err := kafka.NewProducer(kafka.ProducerConfig{
		Brokers:          brokers[0],
		AddressFamily:    cfg.AddressFamily,
		SecurityProtocol: cfg.SecurityProtocol,
		LogLevel:         cfg.LogLevel,
		MaxMessageBytes:  cfg.MaxMessageBytes,
	})
	if err != nil {
		t.Fatalf("creating kafka producer: %s", err)
	}

	t.Cleanup(func() {
		producer.Close()
	})

	return producer
}
