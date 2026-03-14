package payment

import (
	"fmt"
)

// Status represents the status of a payment.
type Status struct {
	value string
}

var paymentStatuses = map[string]bool{
	"pending":   true,
	"succeeded": true,
	"failed":    true,
	"refunded":  true,
}

// Parse parses the string value and returns a PaymentStatus.
func Parse(value string) (Status, error) {
	if !paymentStatuses[value] {
		return Status{}, fmt.Errorf("%w: %s", ErrInvalidPaymentStatus, value)
	}

	return Status{value: value}, nil
}

// MustParse parses the string value. Panics on error.
func MustParse(value string) Status {
	ps, err := Parse(value)
	if err != nil {
		panic(err)
	}

	return ps
}

// String returns the string value.
func (p Status) String() string {
	return p.value
}

// Equal provides support for the go-cmp package and testing.
func (p Status) Equal(p2 Status) bool {
	return p.value == p2.value
}
