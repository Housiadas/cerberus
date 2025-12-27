package auth_usecase

import "errors"

// Specific error variables for auth failures.
var (
	ErrForbidden    = errors.New("attempted action is not allowed")
	ErrKIDMissing   = errors.New("kid missing from token header")
	ErrKIDMalformed = errors.New("kid in token header is malformed")
	ErrUserDisabled = errors.New("user is inactive")
)
