package cursor

import "errors"

var (
	ErrLimitTooSmall  = errors.New("limit value too small, must be larger than 0")
	ErrLimitTooLarge  = errors.New("limit value too large, must be less than or equal to 100")
	ErrInvalidCursor  = errors.New("invalid cursor value")
	ErrInvalidEncoded = errors.New("could not decode cursor")
)
