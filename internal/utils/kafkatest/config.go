package kafkatest

import (
	"github.com/Housiadas/cerberus/internal/config"
)

// Config holds test kafka configuration.
type Config struct {
	AddressFamily    string
	SecurityProtocol string
	LogLevel         int
	MaxMessageBytes  int
	SessionTimeout   int
}

func newConfig(t interface {
	Helper()
	Fatal(args ...any)
},
) Config {
	t.Helper()

	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatal("load kafka config:", err)
	}

	return Config{
		AddressFamily:    cfg.Kafka.AddressFamily,
		SecurityProtocol: cfg.Kafka.SecurityProtocol,
		LogLevel:         cfg.Kafka.LogLevel,
		MaxMessageBytes:  cfg.Kafka.MaxMessageBytes,
		SessionTimeout:   cfg.Kafka.SessionTimeout,
	}
}
