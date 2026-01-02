package config

import "time"

type Rest struct {
	Api             string
	Debug           string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}
