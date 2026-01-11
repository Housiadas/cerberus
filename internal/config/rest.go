package config

import "time"

type Rest struct {
	API             string
	Debug           string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}
