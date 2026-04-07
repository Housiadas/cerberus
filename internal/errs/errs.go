// Package errs provides types and support related to web error functionality.
package errs

import (
	"errors"
	"fmt"
)

// Error represents an error in the system.
type Error struct {
	Status  StatusCode   `json:"status"`
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Fields  []FieldError `json:"fields,omitempty"`
}

// New constructs an error based on an error.
func New(status StatusCode, code string, err error) *Error {
	e := &Error{
		Status:  status,
		Code:    code,
		Message: err.Error(),
	}

	if inner, ok := errors.AsType[*Error](err); ok {
		e.Fields = inner.Fields
	}

	return e
}

// Errorf constructs an error based on an error message.
func Errorf(status StatusCode, code, format string, v ...any) *Error {
	return &Error{
		Status:  status,
		Code:    code,
		Message: fmt.Sprintf(format, v...),
	}
}

// Error implements the error interface.
func (e *Error) Error() string {
	return e.Message
}

// HTTPStatus get the http status code.
func (e *Error) HTTPStatus() int {
	return statusTable[e.Status].http
}

// Equal provides support for the go-cmp package and testing.
func (e *Error) Equal(o *Error) bool {
	return e.Status == o.Status && e.Code == o.Code && e.Message == o.Message
}
