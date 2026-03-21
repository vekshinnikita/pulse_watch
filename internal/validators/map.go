package validators

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/vekshinnikita/pulse_watch/internal/errs"
)

func makeStructuredErrors(key string, err validator.ValidationErrors) []errs.StructuredFieldError {
	var structuredErrors errs.StructuredFieldErrors

	for _, e := range err {
		structuredErrors = append(structuredErrors, errs.StructuredFieldError{
			Field:   key,
			Type:    errs.GetTypeError(e),
			Message: errs.ValidationErrorToText(e),
		})
	}

	return structuredErrors
}

func getMapValidator() *validator.Validate {
	validate, ok := binding.Validator.Engine().(*validator.Validate)
	if ok {
		return validate
	}

	return validator.New()
}

func validateMapWithPrefix(m map[string]any, rules map[string]string, prefix string) errs.StructuredFieldErrors {
	validate := getMapValidator()
	var structuredErrors errs.StructuredFieldErrors

	for key, rule := range rules {
		value, _ := m[key]
		keyWithPrefix := prefix + key
		if err := validate.Var(value, rule); err != nil {
			var ve validator.ValidationErrors
			if errors.As(err, &ve) {
				structuredErrors = append(structuredErrors, makeStructuredErrors(keyWithPrefix, ve)...)
			}
		}
	}

	return structuredErrors
}

func ValidateMap(m map[string]any, rules map[string]string) errs.StructuredFieldErrors {
	return validateMapWithPrefix(m, rules, "")
}

func ValidateListMap(l []map[string]any, rules map[string]string) errs.StructuredFieldErrors {
	var structuredErrors errs.StructuredFieldErrors
	for key, m := range l {
		prefix := fmt.Sprintf("%d.", key)
		structuredErrors = append(structuredErrors, validateMapWithPrefix(m, rules, prefix)...)
	}

	return structuredErrors
}
