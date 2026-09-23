// Package validator wraps go-playground/validator and formats errors for API clients.
package validator

import (
	"errors"  // errors is used for errors.As
	"reflect" // reflect is used to read json tag names
	"strings" // strings is used to split json tags

	"github.com/go-playground/validator/v10" // the actual validation engine
)

// FieldError describes a single invalid field in a request.
type FieldError struct {
	Field   string `json:"field"`   // Field is the JSON field name
	Tag     string `json:"tag"`     // Tag is the rule that failed (required, min, ...)
	Param   string `json:"param"`   // Param is the rule parameter (e.g. "2" for min=2)
	Message string `json:"message"` // Message is a human readable explanation
}

// Validator is a thin wrapper around *validator.Validate.
type Validator struct {
	v *validator.Validate // v is the underlying validator instance
}

// New creates a Validator that reports JSON field names instead of Go field names.
func New() *Validator {
	// Create a new validator instance with required-struct checks enabled
	v := validator.New(validator.WithRequiredStructEnabled())
	// Use the `json` tag as the field name in error messages
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		// Take only the name part of the tag (before any comma)
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		// "-" means the field is ignored by JSON
		if name == "-" {
			return ""
		}
		// Return the JSON name
		return name
	})
	// Return the wrapper
	return &Validator{v: v}
}

// Validate validates a struct and returns a slice of FieldError (nil when valid).
func (val *Validator) Validate(s any) []FieldError {
	// Run the validation rules defined in struct tags
	err := val.v.Struct(s)
	// No error means the struct is valid
	if err == nil {
		return nil
	}
	// Try to convert into validator.ValidationErrors to get per-field details
	var ves validator.ValidationErrors
	if !errors.As(err, &ves) {
		// Unknown error type: return it as a single generic error
		return []FieldError{{Field: "", Tag: "invalid", Message: err.Error()}}
	}
	// Build the list of field errors
	out := make([]FieldError, 0, len(ves))
	for _, fe := range ves {
		out = append(out, FieldError{
			Field:   fe.Field(),  // JSON field name
			Tag:     fe.Tag(),    // failed rule
			Param:   fe.Param(),  // rule parameter
			Message: message(fe), // friendly message
		})
	}
	// Return all field errors
	return out
}

// message builds a friendly message for a single validation error.
func message(fe validator.FieldError) string {
	// Pick a message based on the rule that failed
	switch fe.Tag() {
	case "required":
		return fe.Field() + " is required"
	case "min":
		return fe.Field() + " must be at least " + fe.Param() + " characters"
	case "max":
		return fe.Field() + " must be at most " + fe.Param() + " characters"
	case "gt":
		return fe.Field() + " must be greater than " + fe.Param()
	default:
		return fe.Field() + " is invalid"
	}
}
