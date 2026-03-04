package config

import "time"

type Collector struct {
	Host string
	// Shouldn't use a high Probability value in non-developer systems.
	// 0.05 should be enough for most systems. Some might want to have
	// this even lower.
	Probability float64
	// MetricInterval controls how often metrics are exported. Default is 30s.
	MetricInterval time.Duration
}
