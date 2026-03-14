// Package currency represents currency in the system.
package currency

import (
	"fmt"
	"regexp"
)

// Currency represents a currency code.
type Currency struct {
	value string
}

var currencyRegex = regexp.MustCompile(`^[A-Z]{3}$`)

// Parse parses the string value and returns a Currency.
func Parse(value string) (Currency, error) {
	if !currencyRegex.MatchString(value) {
		return Currency{}, fmt.Errorf("%w: %s", ErrInvalidCurrency, value)
	}

	return Currency{value: value}, nil
}

// MustParse parses the string value. Panics on error.
func MustParse(value string) Currency {
	c, err := Parse(value)
	if err != nil {
		panic(err)
	}

	return c
}

// String returns the string value.
func (c Currency) String() string {
	return c.value
}

// Equal provides support for the go-cmp package and testing.
func (c Currency) Equal(c2 Currency) bool {
	return c.value == c2.value
}
