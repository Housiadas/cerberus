// Package validation contains the support for validating models.
package validation

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/Housiadas/cerberus/internal/errs"
	"github.com/go-playground/validator/v10"
)

// validate holds the settings and caches for validating request struct values.
var validate = newValidator()

func newValidator() *validator.Validate {
	v := validator.New(validator.WithRequiredStructEnabled())

	// Use JSON tag names for errors instead of Go struct names.
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}

		return name
	})

	return v
}

// Check validates the provided model against its declared tags.
func Check(val any) error {
	err := validate.Struct(val)
	if err != nil {
		var vErrors validator.ValidationErrors

		ok := errors.As(err, &vErrors)
		if !ok {
			return fmt.Errorf("validation error: %w", err)
		}

		var fields errs.FieldErrors

		for _, verror := range vErrors {
			msg := formatValidationError(verror)
			fields.Add(
				verror.Field(),
				//nolint:err113 // Dynamic validation messages are required
				errors.New(msg),
			)
		}

		return &fields
	}

	return nil
}

func formatValidationError(ve validator.FieldError) string {
	field := strings.ToLower(ve.Field())

	switch ve.Tag() {
	case "required":
		return field + " is required"
	case "email":
		return field + " must be a valid email"
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, ve.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", field, ve.Param())
	default:
		return fmt.Sprintf("%s failed validation on '%s'", field, ve.Tag())
	}
}
