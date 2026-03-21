package gin_validators

import (
	"fmt"
	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/vekshinnikita/pulse_watch/internal/constants"
)

var enumValidator validator.Func = func(fl validator.FieldLevel) bool {
	param := fl.Param()
	value := fl.Field().Interface()

	enumSet, ok := constants.EnumSetsMap[param]
	if !ok {
		slog.Warn(fmt.Sprintf("unknown enum set name: %s", param))
		return true
	}

	_, ok = enumSet[value]
	if !ok {
		return false
	}

	return true
}
