package testutils

import (
	"io"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/vekshinnikita/pulse_watch/internal/config"
	gin_validators "github.com/vekshinnikita/pulse_watch/internal/validators/gin"
)

func BaseSetup() {
	config.LoadConfig(os.Getenv("CONFIG_PATH"))
	godotenv.Load(os.Getenv("ENV_FILE_PATH"))

	// Отключаем логгер
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func HandlerSetup() {
	BaseSetup()

	// Убираем вывод информации от gin
	gin.SetMode(gin.ReleaseMode)

	// Инициализируем кастомные валидаторы
	_ = gin_validators.InitValidators()
}

func ServiceSetup() {
	BaseSetup()
}

func RepositorySetup() {
	BaseSetup()
}
