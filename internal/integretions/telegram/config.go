package telegram_integration

import (
	"log/slog"
	"os"
	"sync"
)

var (
	onceConfig sync.Once
	cfg        *Config
)

type Config struct {
	ApiToken string
}

func GetConfig() *Config {
	onceConfig.Do(func() {
		cfg = &Config{
			ApiToken: os.Getenv("TELEGRAM_API_TOKEN"),
		}
	})

	if cfg == nil {
		slog.Error("common config not loaded")
	}

	return cfg
}
