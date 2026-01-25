package web

import "errors"

var (
	ErrPageTooSmall       = errors.New("page value too small, must be larger than 0")
	ErrRowsTooSmall       = errors.New("rows value too small, must be larger than 0")
	ErrRowsTooLarge       = errors.New("rows value too large, must be less than 100")
	ErrClientDisconnected = errors.New("client disconnected, do not send encode")
)
