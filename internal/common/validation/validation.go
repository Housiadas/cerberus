// Package validation contains the support for validating models.
package validation

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/Housiadas/cerberus/pkg/web/errs"
)

// Check validates the provided model against it's declared tags.
func Check(val any) error {
	if err := validate.Struct(val); err != nil {
		var vErrors validator.ValidationErrors
		ok := errors.As(err, &vErrors)
		if !ok {
			return err
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
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, ve.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", field, ve.Param())
	default:
		return fmt.Sprintf("%s failed validation on '%s'", field, ve.Tag())
	}
}
