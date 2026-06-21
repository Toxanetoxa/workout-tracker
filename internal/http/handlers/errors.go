package handlers

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/go-playground/validator/v10"
)

func validationErrors(err error) []ValidationErrorItem {
	return validationErrorsWithField("", err)
}

func validationErrorsWithField(fieldName string, err error) []ValidationErrorItem {
	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) || len(validationErrors) == 0 {
		return []ValidationErrorItem{{Field: fieldName, Rule: "invalid", Message: "validation failed"}}
	}

	slices.SortFunc(validationErrors, func(a, b validator.FieldError) int {
		return strings.Compare(validationFieldName(a), validationFieldName(b))
	})

	items := make([]ValidationErrorItem, 0, len(validationErrors))
	for _, validationErr := range validationErrors {
		items = append(items, validationErrorItem(fieldName, validationErr))
	}

	return items
}

type ValidationErrorResponse struct {
	Errors []ValidationErrorItem `json:"errors"`
}

type ValidationErrorItem struct {
	Field   string `json:"field"`
	Rule    string `json:"rule"`
	Message string `json:"message"`
}

func validationErrorItem(fieldName string, err validator.FieldError) ValidationErrorItem {
	field := validationFieldName(err)
	if fieldName != "" {
		field = fieldName
	}

	switch err.Tag() {
	case "required":
		return ValidationErrorItem{Field: field, Rule: err.Tag(), Message: fmt.Sprintf("%s is required", field)}
	case "min":
		return ValidationErrorItem{Field: field, Rule: err.Tag(), Message: fmt.Sprintf("%s must be at least %s characters long", field, err.Param())}
	case "max":
		return ValidationErrorItem{Field: field, Rule: err.Tag(), Message: fmt.Sprintf("%s must be at most %s characters long", field, err.Param())}
	case "gt":
		return ValidationErrorItem{Field: field, Rule: err.Tag(), Message: fmt.Sprintf("%s must be greater than %s", field, err.Param())}
	case "trimmed":
		return ValidationErrorItem{Field: field, Rule: err.Tag(), Message: fmt.Sprintf("%s must not contain leading or trailing spaces", field)}
	case "alphanumdash":
		return ValidationErrorItem{Field: field, Rule: err.Tag(), Message: fmt.Sprintf("%s must contain only letters, digits, dashes and underscores", field)}
	default:
		return ValidationErrorItem{Field: field, Rule: err.Tag(), Message: fmt.Sprintf("%s is invalid", field)}
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
