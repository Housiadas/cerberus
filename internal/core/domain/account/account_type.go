package account

import (
	"fmt"
)

// Type represents the type of a billing account.
type Type struct {
	value string
}

var accountTypes = map[string]bool{
	"personal":   true,
	"team":       true,
	"enterprise": true,
}

// Parse parses the string value and returns an AccountType.
func Parse(value string) (Type, error) {
	if !accountTypes[value] {
		return Type{}, fmt.Errorf("%w: %s", ErrInvalidAccountType, value)
	}

	return Type{value: value}, nil
}

// MustParse parses the string value. Panics on error.
func MustParse(value string) Type {
	at, err := Parse(value)
	if err != nil {
		panic(err)
	}

	return at
}

// String returns the string value.
func (a Type) String() string {
	return a.value
}

// Equal provides support for the go-cmp package and testing.
func (a Type) Equal(a2 Type) bool {
	return a.value == a2.value
}
