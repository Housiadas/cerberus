package subscription

import (
	"fmt"
)

// Status represents the status of a subscription.
type Status struct {
	value string
}

var subscriptionStatuses = map[string]bool{
	"trialing":  true,
	"active":    true,
	"past_due":  true,
	"cancelled": true,
	"unpaid":    true,
}

// Parse parses the string value and returns a SubscriptionStatus.
func Parse(value string) (Status, error) {
	if !subscriptionStatuses[value] {
		return Status{}, fmt.Errorf("%w: %s", ErrInvalidSubscriptionStatus, value)
	}

	return Status{value: value}, nil
}

// MustParse parses the string value. Panics on error.
func MustParse(value string) Status {
	ss, err := Parse(value)
	if err != nil {
		panic(err)
	}

	return ss
}

// String returns the string value.
func (s Status) String() string {
	return s.value
}

// Equal provides support for the go-cmp package and testing.
func (s Status) Equal(s2 Status) bool {
	return s.value == s2.value
}
