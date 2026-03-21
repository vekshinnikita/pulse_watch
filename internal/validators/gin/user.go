package gin_validators

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

var passwordValidator validator.Func = func(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	// Минимум 8 символов, хотя можно сделать сложнее
	if len(password) < 8 {
		return false
	}

	// Проверка на хотя бы одну цифру
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
	if !hasDigit {
		return false
	}

	// Проверка на хотя бы одну букву
	hasLetter := regexp.MustCompile(`[A-Za-z]`).MatchString(password)
	if !hasLetter {
		return false
	}

	return true
}
