package plan

import (
	"fmt"
)

// Interval represents the interval of a billing plan.
type Interval struct {
	value string
}

var planIntervals = map[string]bool{
	"monthly": true,
	"yearly":  true,
}

// Parse parses the string value and returns a PlanInterval.
func Parse(value string) (Interval, error) {
	if !planIntervals[value] {
		return Interval{}, fmt.Errorf("%w: %s", ErrInvalidPlanInterval, value)
	}

	return Interval{value: value}, nil
}

// MustParse parses the string value. Panics on error.
func MustParse(value string) Interval {
	pi, err := Parse(value)
	if err != nil {
		panic(err)
	}

	return pi
}

// String returns the string value.
func (p Interval) String() string {
	return p.value
}

// Equal provides support for the go-cmp package and testing.
func (p Interval) Equal(p2 Interval) bool {
	return p.value == p2.value
}
