package validation

import (
	"regexp"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

var Validate *validator.Validate

func init() {
	// Инициализируем валидатор
	Validate = validator.New()

	// Регистрируем валидацию для username
	Validate.RegisterValidation("username", validateUsername)

	// Подключаем валидацию к Gin
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("username", validateUsername)
	}
}

// validateUsername функция валидации username
func validateUsername(fl validator.FieldLevel) bool {
	pattern := "^[a-zA-Z][a-zA-Z0-9]{5,14}$"
	match, _ := regexp.MatchString(pattern, fl.Field().String())
	return match
}
