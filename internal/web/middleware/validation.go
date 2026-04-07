package middleware

import (
	"context"
	"net/http"
	"reflect"

	"github.com/Housiadas/cerberus/internal/sdk/errs"
	"github.com/Housiadas/cerberus/internal/sdk/validation"
	"github.com/Housiadas/cerberus/internal/web/handler/openapi"
)

// Validation validates request bodies using struct validation tags.
func (m *Middleware) Validation() openapi.StrictMiddlewareFunc {
	return func(
		f openapi.StrictHandlerFunc,
		_ string,
	) openapi.StrictHandlerFunc {
		return func(
			ctx context.Context,
			w http.ResponseWriter,
			r *http.Request,
			request any,
		) (any, error) {
			body := extractBody(request)
			if body == nil {
				return f(ctx, w, r, request)
			}

			err := validation.Check(body)
			if err != nil {
				return nil, errs.ParseValidationErrors(err)
			}

			return f(ctx, w, r, request)
		}
	}
}

// extractBody extracts the Body field from a request object.
func extractBody(request any) any {
	v := reflect.ValueOf(request)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil
	}

	bodyField := v.FieldByName("Body")
	if !bodyField.IsValid() || bodyField.IsNil() {
		return nil
	}

	return bodyField.Interface()
}
