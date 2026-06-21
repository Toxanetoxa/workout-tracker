package handlers

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

func validationMessage(err error) string {
	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) || len(validationErrors) == 0 {
		return "validation failed"
	}

	first := validationErrors[0]
	field := validationFieldName(first)

	switch first.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", field, first.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters long", field, first.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, first.Param())
	case "trimmed":
		return fmt.Sprintf("%s must not contain leading or trailing spaces", field)
	case "alphanumdash":
		return fmt.Sprintf("%s must contain only letters, digits, dashes and underscores", field)
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}

func validationFieldName(err validator.FieldError) string {
	switch err.StructField() {
	case "UserID":
		return "user_id"
	case "ExerciseID":
		return "exercise_id"
	case "PerformedAt":
		return "performed_at"
	case "Name":
		return "name"
	default:
		return err.Field()
	}
}
