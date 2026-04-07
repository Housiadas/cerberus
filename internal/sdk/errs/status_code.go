package errs

import (
	"errors"
	"fmt"
)

var ErrCodeNotExist = errors.New("err code does not exist")

// StatusCode represents an HTTP status category in the system.
type StatusCode int

// Value returns the integer value of the status code.
func (sc *StatusCode) Value() int {
	return int(*sc)
}

// String returns the string representation of the status code.
func (sc *StatusCode) String() string {
	return statusTable[*sc].name
}

// UnmarshalText implement the unmarshal interface for JSON conversions.
func (sc *StatusCode) UnmarshalText(data []byte) error {
	v, ok := statusByName[string(data)]
	if !ok {
		return fmt.Errorf("%w: %q", ErrCodeNotExist, data)
	}

	*sc = v

	return nil
}

// MarshalText implement the marshal interface for JSON conversions.
func (sc *StatusCode) MarshalText() ([]byte, error) {
	return []byte(sc.String()), nil
}

// Equal provides support for the go-cmp package and testing.
func (sc *StatusCode) Equal(o StatusCode) bool {
	return *sc == o
}
