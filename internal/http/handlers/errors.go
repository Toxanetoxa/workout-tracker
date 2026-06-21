package handlers

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/go-playground/validator/v10"
)

func validationMessage(err error) string {
	return validationMessageWithField("", err)
}

func validationMessageWithField(fieldName string, err error) string {
	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) || len(validationErrors) == 0 {
		return "validation failed"
	}

	slices.SortFunc(validationErrors, func(a, b validator.FieldError) int {
		return strings.Compare(validationFieldName(a), validationFieldName(b))
	})

	messages := make([]string, 0, len(validationErrors))
	for _, validationErr := range validationErrors {
		messages = append(messages, validationErrorMessage(fieldName, validationErr))
	}

	return strings.Join(messages, "; ")
}

func validationErrorMessage(fieldName string, err validator.FieldError) string {
	field := validationFieldName(err)
	if fieldName != "" {
		field = fieldName
	}

	switch err.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", field, err.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters long", field, err.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, err.Param())
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
