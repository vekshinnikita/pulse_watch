package gin_validators

import (
	"time"

	"github.com/go-playground/validator/v10"
)

var futureTimeValidator validator.Func = func(fl validator.FieldLevel) bool {
	value := fl.Field()

	date, ok := value.Interface().(time.Time)
	if ok {
		return date.After(time.Now())
	}

	return false
}

var typeValidator validator.Func = func(fl validator.FieldLevel) bool {
	param := fl.Param()
	value := fl.Field().Interface()

	switch param {
	case "string":
		_, ok := value.(string)
		return ok
	case "int":
		_, ok := value.(int)
		return ok
	default:
		return false
	}
}
