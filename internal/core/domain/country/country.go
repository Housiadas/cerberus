// Package country represents country in the system.
package country

import (
	"fmt"
	"regexp"
)

// Country represents a country code (ISO 3166-1 alpha-2).
type Country struct {
	value string
}

var countryRegex = regexp.MustCompile(`^[A-Z]{2}$`)

// Parse parses the string value and returns a Country.
func Parse(value string) (Country, error) {
	if !countryRegex.MatchString(value) {
		return Country{}, fmt.Errorf("%w: %s", ErrInvalidCountry, value)
	}

	return Country{value: value}, nil
}

// MustParse parses the string value. Panics on error.
func MustParse(value string) Country {
	c, err := Parse(value)
	if err != nil {
		panic(err)
	}

	return c
}

// String returns the string value.
func (c Country) String() string {
	return c.value
}

// Equal provides support for the go-cmp package and testing.
func (c Country) Equal(c2 Country) bool {
	return c.value == c2.value
}
