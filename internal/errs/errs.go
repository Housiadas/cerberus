// Package errs provides types and support related to web error functionality.
package errs

import (
	"errors"
	"fmt"
	"runtime"
)

// Error represents an error in the system.
type Error struct {
	Status   StatusCode   `json:"status"`
	Code     string       `json:"code"`
	Message  string       `json:"message"`
	Fields   []FieldError `json:"fields,omitempty"`
	FuncName string       `json:"-"`
	FileName string       `json:"-"`
}

// New constructs an error based on an error.
func New(status StatusCode, code string, err error) *Error {
	pc, filename, line, _ := runtime.Caller(1)

	e := &Error{
		Status:   status,
		Code:     code,
		Message:  err.Error(),
		FuncName: runtime.FuncForPC(pc).Name(),
		FileName: fmt.Sprintf("%s:%d", filename, line),
	}

	if inner, ok := errors.AsType[*Error](err); ok {
		e.Fields = inner.Fields
	}

	return e
}

// Errorf constructs an error based on an error message.
func Errorf(status StatusCode, code string, format string, v ...any) *Error {
	pc, filename, line, _ := runtime.Caller(1)

	return &Error{
		Status:   status,
		Code:     code,
		Message:  fmt.Sprintf(format, v...),
		FuncName: runtime.FuncForPC(pc).Name(),
		FileName: fmt.Sprintf("%s:%d", filename, line),
	}
}

// Error implements the error interface.
func (e *Error) Error() string {
	return e.Message
}

// HTTPStatus get the http status code.
func (e *Error) HTTPStatus() int {
	return httpStatus[e.Status]
}

// Equal provides support for the go-cmp package and testing.
func (e *Error) Equal(e2 *Error) bool {
	return e.Status == e2.Status && e.Code == e2.Code && e.Message == e2.Message
}
