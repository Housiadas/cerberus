package password

import "errors"

var (
	errInvalidPass = errors.New("invalid password")
	errPassNoMatch = errors.New("passwords do not match")
)
