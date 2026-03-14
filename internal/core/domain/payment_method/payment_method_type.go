package payment_method

import (
	"fmt"
)

// PaymentMethodType represents the type of a payment method.
type PaymentMethodType struct {
	value string
}

var paymentMethodTypes = map[string]bool{
	"card":          true,
	"bank_transfer": true,
	"sepa":          true,
	"paypal":        true,
}

// Parse parses the string value and returns a PaymentMethodType.
func Parse(value string) (PaymentMethodType, error) {
	if !paymentMethodTypes[value] {
		return PaymentMethodType{}, fmt.Errorf("%w: %s", ErrInvalidPaymentMethodType, value)
	}

	return PaymentMethodType{value: value}, nil
}

// MustParse parses the string value. Panics on error.
func MustParse(value string) PaymentMethodType {
	pmt, err := Parse(value)
	if err != nil {
		panic(err)
	}

	return pmt
}

// String returns the string value.
func (p PaymentMethodType) String() string {
	return p.value
}

// Equal provides support for the go-cmp package and testing.
func (p PaymentMethodType) Equal(p2 PaymentMethodType) bool {
	return p.value == p2.value
}
