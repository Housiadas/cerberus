package errs

import (
	"errors"
	"fmt"
	"runtime"
)

func ParseValidationErrors(err error) *Error {
	var fieldErrors *FieldErrors

	ok := errors.As(err, &fieldErrors)
	if !ok {
		return Errorf(InvalidArgument, "validation error: %s", err.Error())
	}

	feSlice := toFieldErrorSlice(*fieldErrors)

	return newWithFields(InvalidArgument, errors.New("validation error"), feSlice)
}

func newWithFields(code ErrCode, err error, fe []FieldError) *Error {
	pCounter, filename, line, _ := runtime.Caller(1)

	return &Error{
		Code:     code,
		Message:  err.Error(),
		Fields:   fe,
		FuncName: runtime.FuncForPC(pCounter).Name(),
		FileName: fmt.Sprintf("%s:%d", filename, line),
	}
}

func toFieldErrorSlice(fe FieldErrors) []FieldError {
	var result []FieldError
	for _, err := range fe {
		result = append(result, FieldError{
			Field: err.Field,
			Err:   err.Err,
		})
	}

	return result
}
