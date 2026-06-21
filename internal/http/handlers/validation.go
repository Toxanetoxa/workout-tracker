package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

var alphanumDashPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]*$`)

func registerValidationRules(v *validator.Validate) {
	_ = v.RegisterValidation("trimmed", func(fl validator.FieldLevel) bool {
		value, ok := fl.Field().Interface().(string)
		if !ok {
			return false
		}

		return strings.TrimSpace(value) == value && value != ""
	})

	_ = v.RegisterValidation("alphanumdash", func(fl validator.FieldLevel) bool {
		value, ok := fl.Field().Interface().(string)
		if !ok {
			return false
		}

		return alphanumDashPattern.MatchString(value)
	})
}

func decodeJSONBody(r *http.Request, dst any) error {
	limited := io.LimitReader(r.Body, 1<<20)
	dec := json.NewDecoder(limited)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		return err
	}

	var extra any
	if err := dec.Decode(&extra); err != nil && !errors.Is(err, io.EOF) {
		return err
	}

	return nil
}
