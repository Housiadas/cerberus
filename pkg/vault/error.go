package vault

import "errors"

var (
	ErrCreateClient       = errors.New("failed to create vault client")
	ErrSetToken           = errors.New("failed to set vault token")
	ErrReadSecret         = errors.New("failed to read secret")
	ErrGetJWTSecret       = errors.New("failed to get jwt secret")
	ErrInvalidSecretValue = errors.New("access_token_secret not found or invalid type")
)
