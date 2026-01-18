// Package password represents a password in the system.
package password

// Password represents a password in the system.
type Password struct {
	value string
}

// String returns the value of the password.
func (n Password) String() string {
	return n.value
}

// Equal provides support for the go-cmp package and testing.
func (n Password) Equal(n2 Password) bool {
	return n.value == n2.value
}

// MarshalText provides support for logging and any marshal needs.
func (n Password) MarshalText() ([]byte, error) {
	return []byte(n.value), nil
}
