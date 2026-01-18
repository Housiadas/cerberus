package middleware

import "errors"

var (
	ErrInvalidAuthHeader = errors.New("expected authorization header format: Bearer <token>")
	ErrInvalidBasicAuth  = errors.New("invalid Basic auth")
	ErrPermissionDenied  = errors.New("permission denied")
)
