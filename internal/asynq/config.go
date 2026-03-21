package asynq_server

import (
	"log/slog"
	"sync"

	"github.com/spf13/viper"
)

type BackgroundTasksConfig []struct {
	Task string `mapstructure:"task"`
	Cron string `mapstructure:"cron"`
}

type InitTasks []string

var (
	onceBackgroundConfig sync.Once
	onceInitConfig       sync.Once
	backgroundCfg        BackgroundTasksConfig
	initTasks            InitTasks
)

func GetBackgroundTasks() BackgroundTasksConfig {
	onceBackgroundConfig.Do(func() {
		viper.UnmarshalKey("tasks", &backgroundCfg)
	})

	if backgroundCfg == nil {
		slog.Error("common config not loaded")
	}

	return backgroundCfg
}
