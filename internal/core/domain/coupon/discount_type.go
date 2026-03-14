package coupon

import (
	"fmt"
)

// DiscountType represents the type of a discount.
type DiscountType struct {
	value string
}

var discountTypes = map[string]bool{
	"percent": true,
	"fixed":   true,
}

// Parse parses the string value and returns a DiscountType.
func Parse(value string) (DiscountType, error) {
	if !discountTypes[value] {
		return DiscountType{}, fmt.Errorf("%w: %s", ErrInvalidDiscountType, value)
	}

	return DiscountType{value: value}, nil
}

// MustParse parses the string value. Panics on error.
func MustParse(value string) DiscountType {
	dt, err := Parse(value)
	if err != nil {
		panic(err)
	}

	return dt
}

// String returns the string value.
func (d DiscountType) String() string {
	return d.value
}

// Equal provides support for the go-cmp package and testing.
func (d DiscountType) Equal(d2 DiscountType) bool {
	return d.value == d2.value
}
