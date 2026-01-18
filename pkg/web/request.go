package web

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Param returns the web call parameters from the request.
func Param(r *http.Request, key string) string {
	return r.PathValue(key)
}

// Header returns the web call headers from the request.
func Header(r *http.Request, key string) string {
	return r.Header.Get(key)
}

type validator interface {
	Validate() error
}

// Decode reads the body of an HTTP request looking for a JSON document.
// The body is decoded into the provided value.
func Decode(r *http.Request, val any) error {
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()

	err := d.Decode(val)
	if err != nil {
		return fmt.Errorf("unable to decode payload: %w", err)
	}

	// If the provided value is a struct, then it is checked for validation tags.
	// If the value implements a validate function, it is executed.
	if v, ok := val.(validator); ok {
		err := v.Validate()
		if err != nil {
			return fmt.Errorf("web decode validation: %w", err)
		}
	}

	return nil
}
