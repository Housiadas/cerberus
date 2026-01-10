// Package web contains the logic for the web
package web

import (
	"context"
	"net/http"
)

// HandlerFunc represents a function that handles an http request.
type HandlerFunc func(ctx context.Context, w http.ResponseWriter, r *http.Request) Encoder

// Encoder defines behavior that can encode a data model and provide the content type for that encoding.
type Encoder interface {
	Encode() (data []byte, contentType string, err error)
}

type httpStatus interface {
	HTTPStatus() int
}
