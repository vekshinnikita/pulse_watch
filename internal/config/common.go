package config

import (
	"log/slog"
	"sync"

	"github.com/spf13/viper"
)

var (
	onceCommonConfig sync.Once
	commonCfg        *CommonConfig
)

type CommonConfig struct {
	Debug bool
}

func GetCommonConfig() *CommonConfig {
	onceCommonConfig.Do(func() {
		// Путь до файла
		debug := viper.GetBool("debug")

		commonCfg = &CommonConfig{
			Debug: debug,
		}
	})

	if commonCfg == nil {
		slog.Error("common config not loaded")
	}

	return commonCfg
}
