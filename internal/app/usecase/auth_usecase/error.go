package auth_usecase

import "errors"

var (
	ErrUserDisabled = errors.New("user is inactive")
)
