package errs

import (
	"errors"
	"fmt"
)

var ErrCodeNotExist = errors.New("err code does not exist")

// StatusCode represents an HTTP status category in the system.
type StatusCode struct {
	value int
}

// Value returns the integer value of the status code.
func (sc *StatusCode) Value() int {
	return sc.value
}

// String returns the string representation of the status code.
func (sc *StatusCode) String() string {
	return statusNames[*sc]
}

// UnmarshalText implement the unmarshal interface for JSON conversions.
func (sc *StatusCode) UnmarshalText(data []byte) error {
	name := string(data)

	v, exists := statusNumbers[name]
	if !exists {
		return fmt.Errorf("%w: %q", ErrCodeNotExist, name)
	}

	*sc = v

	return nil
}

// MarshalText implement the marshal interface for JSON conversions.
func (sc *StatusCode) MarshalText() ([]byte, error) {
	return []byte(sc.String()), nil
}

// Equal provides support for the go-cmp package and testing.
func (sc *StatusCode) Equal(sc2 StatusCode) bool {
	return sc.value == sc2.value
}
