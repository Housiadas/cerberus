package invoice

import (
	"fmt"
)

// Status represents the status of an invoice.
type Status struct {
	value string
}

var invoiceStatuses = map[string]bool{
	"draft":         true,
	"open":          true,
	"paid":          true,
	"void":          true,
	"uncollectible": true,
}

// Parse parses the string value and returns an InvoiceStatus.
func Parse(value string) (Status, error) {
	if !invoiceStatuses[value] {
		return Status{}, fmt.Errorf("%w: %s", ErrInvalidInvoiceStatus, value)
	}

	return Status{value: value}, nil
}

// MustParse parses the string value. Panics on error.
func MustParse(value string) Status {
	is, err := Parse(value)
	if err != nil {
		panic(err)
	}

	return is
}

// String returns the string value.
func (i Status) String() string {
	return i.value
}

// Equal provides support for the go-cmp package and testing.
func (i Status) Equal(i2 Status) bool {
	return i.value == i2.value
}
