package logger

import (
	"log/slog"
	"sync"

	"github.com/spf13/viper"
)

type LoggerConfig struct {
	Filepath      string
	MaxAgeDays    int
	RotationHours int
}

var (
	cfg  *LoggerConfig
	once sync.Once
)

func GetConfig() *LoggerConfig {
	once.Do(func() {
		// Путь до файла
		filepath := viper.GetString("logger.filepath")
		if filepath == "" {
			filepath = "logs/%Y-%m-%d.log"
		}

		// Время хранения логов (в днях)
		maxAgeDays := viper.GetInt("logger.max_age_days")
		if maxAgeDays == 0 {
			maxAgeDays = 30
		}

		// Время ротации лога (в часах)
		rotationHours := viper.GetInt("logger.rotation_hours")
		if rotationHours == 0 {
			rotationHours = 24
		}

		cfg = &LoggerConfig{
			Filepath:      filepath,
			MaxAgeDays:    maxAgeDays,
			RotationHours: rotationHours,
		}
	})

	if cfg == nil {
		slog.Error("server config not loaded")
	}

	return cfg
}
