package gin_validators

import (
	"fmt"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func InitValidators() error {

	// Регистрируем кастомную валидацию
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return fmt.Errorf("error while getting validator engine")
	}

	v.RegisterValidation("password", passwordValidator)
	v.RegisterValidation("future", futureTimeValidator)
	v.RegisterValidation("type", typeValidator)
	v.RegisterValidation("enum", enumValidator)

	return nil
}
