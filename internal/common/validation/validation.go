// Package validation contains the support for validating models.
package validation

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Housiadas/cerberus/pkg/web/errs"
	"github.com/go-playground/validator/v10"
)

// Check validates the provided model against it's declared tags.
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
			// Create a human-readable message
			msg := formatValidationError(verror)
			fields.Add(
				verror.Field(),
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
