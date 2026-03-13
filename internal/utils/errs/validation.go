package errs

import (
	"errors"
	"fmt"
	"runtime"
)

var ErrValidationError = errors.New("validation error")

// ParseValidationErrors converts validation errors into an Error.
func ParseValidationErrors(err error) *Error {
	var fieldErrors *FieldErrors

	ok := errors.As(err, &fieldErrors)
	if !ok {
		return Errorf(InvalidArgument, CodeValidation, "validation error: %s", err.Error())
	}

	feSlice := toFieldErrorSlice(*fieldErrors)

	return newWithFields(InvalidArgument, CodeValidation, ErrValidationError, feSlice)
}

func newWithFields(status StatusCode, code string, err error, fe []FieldError) *Error {
	pCounter, filename, line, _ := runtime.Caller(1)

	return &Error{
		Status:   status,
		Code:     code,
		Message:  err.Error(),
		Fields:   fe,
		FuncName: runtime.FuncForPC(pCounter).Name(),
		FileName: fmt.Sprintf("%s:%d", filename, line),
	}
}

func toFieldErrorSlice(fe FieldErrors) []FieldError {
	result := make([]FieldError, 0, len(fe))
	for _, err := range fe {
		result = append(result, FieldError{
			Field: err.Field,
			Err:   err.Err,
		})
	}

	return result
}
