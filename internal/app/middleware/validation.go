package middleware

import (
	"context"
	"net/http"
	"reflect"

	"github.com/Housiadas/cerberus/internal/app/handler/openapi"
	"github.com/Housiadas/cerberus/internal/utils/errs"
)

type validator interface {
	Validate() error
}

// Validation validates request bodies that implement the validator interface.
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
			if v := extractValidator(request); v != nil {
				err := v.Validate()
				if err != nil {
					return nil, errs.ParseValidationErrors(err)
				}
			}

			return f(ctx, w, r, request)
		}
	}
}

// extractValidator extracts the Body field from a request object and checks
// if it implements the validator interface.
func extractValidator(request any) validator {
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

	if val, ok := bodyField.Interface().(validator); ok {
		return val
	}

	return nil
}
