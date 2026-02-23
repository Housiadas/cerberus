package errs

import (
	"encoding/json"
)

// FieldErrors represents a collection of field errors.
type FieldErrors []FieldError

// FieldError is used to indicate an error with a specific request field.
type FieldError struct {
	Field string `json:"field"`
	Err   string `json:"error"`
}

// NewFieldErrors creates a field error.
func NewFieldErrors(field string, err error) *Error {
	fe := FieldErrors{
		{
			Field: field,
			Err:   err.Error(),
		},
	}

	return fe.ToError()
}

// Add adds a field error to the collection.
func (fe *FieldErrors) Add(field string, err error) {
	*fe = append(*fe, FieldError{
		Field: field,
		Err:   err.Error(),
	})
}

// ToError converts the field errors to an Error.
func (fe *FieldErrors) ToError() *Error {
	return New(InvalidArgument, fe)
}

// Error implements the error interface.
func (fe *FieldErrors) Error() string {
	d, err := json.Marshal(fe)
	if err != nil {
		return err.Error()
	}

	return string(d)
}
