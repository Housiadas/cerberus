package kafkatest

import (
	"testing"

	"github.com/Housiadas/cerberus/internal/config"
	"github.com/stretchr/testify/require"
)

type Config struct {
	AddressFamily    string
	SecurityProtocol string
	LogLevel         int
	MaxMessageBytes  int
	SessionTimeout   int
}

func newConfig(t *testing.T) Config {
	t.Helper()

	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	return Config{
		AddressFamily:    cfg.Kafka.AddressFamily,
		SecurityProtocol: cfg.Kafka.SecurityProtocol,
		LogLevel:         cfg.Kafka.LogLevel,
		MaxMessageBytes:  cfg.Kafka.MaxMessageBytes,
		SessionTimeout:   cfg.Kafka.SessionTimeout,
	}
}
