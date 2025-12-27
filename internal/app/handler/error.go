package handler

import "errors"

// ErrInvalidID represents a condition where the id is not an uuid.
var ErrInvalidID = errors.New("ID is not in its proper form")
