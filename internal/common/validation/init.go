package validation

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// validate holds the settings and caches for validating request struct values.
var validate *validator.Validate

//nolint:gochecknoinits
func init() {
	// Instantiate a validator.
	validate = validator.New(validator.WithRequiredStructEnabled())

	// Use JSON tag names for errors instead of Go struct names.
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}

		return name
	})
}
