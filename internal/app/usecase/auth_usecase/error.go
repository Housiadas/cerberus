package auth_usecase

import "errors"

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrUserDisabled = errors.New("user is inactive")
	ErrExpiredToken = errors.New("token has expired")
)
