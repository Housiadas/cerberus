package errs

import "errors"

var (
	ErrCodeNotExist    = errors.New("err code does not exist")
	ErrValidationError = errors.New("validation error")
)
