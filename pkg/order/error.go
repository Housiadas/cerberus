package order

import "errors"

var (
	ErrUnknownOrder     = errors.New("unknown order")
	ErrUnknownDirection = errors.New("unknown direction")
)
