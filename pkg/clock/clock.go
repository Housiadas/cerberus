package clock

import "time"

// Clock is an abstraction for time.Time package that allows for testing.
type Clock interface {
	Now() time.Time
	Since(t time.Time) time.Duration
}

// RealClock provides functions that wraps the real time.Time package.
type RealClock struct{}

// NewClock returns a new RealClock.
func NewClock() Clock {
	return &RealClock{}
}

// Now wraps time.Now() from the standard library.
func (c *RealClock) Now() time.Time {
	return time.Now()
}

// Since wraps time.Since() from the standard library.
func (c *RealClock) Since(t time.Time) time.Duration {
	return time.Since(t)
}
