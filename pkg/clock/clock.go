package clock

import "time"

// RealClock provides functions that wrap the real time.Time package.
type RealClock struct{}

// NewClock returns a new RealClock.
func NewClock() *RealClock {
	return &RealClock{}
}

// Now wraps time.Now() from the standard library.
func (c *RealClock) Now() time.Time {
	return time.Now()
}
