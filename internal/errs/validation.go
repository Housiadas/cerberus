package errs

import "errors"

var ErrValidationError = errors.New("validation error")

// ParseValidationErrors converts validation errors into an Error.
func ParseValidationErrors(err error) *Error {
	fe, ok := errors.AsType[*FieldErrors](err)
	if ok {
		return &Error{
			Status:  InvalidArgument,
			Code:    CodeValidation,
			Message: ErrValidationError.Error(),
			Fields:  *fe,
		}
	}

	return Errorf(InvalidArgument, CodeValidation, "validation error: %s", err.Error())
}
