package utils

import (
	"errors"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Converts validator errors into clean UI messages
func ParseValidationError(err error) string {

	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		for _, fe := range ve {

			field := strings.ToLower(fe.Field())

			switch fe.Tag() {
			case "required":
				return field + " is required"

			case "email":
				return "invalid email format"

			case "min":
				return field + " must be at least " + fe.Param() + " characters"

			default:
				return field + " is invalid"
			}
		}
	}
	// default - return same
	return err.Error()
}
