package queue_handlers

import (
	"log/slog"
	"sync"

	"github.com/spf13/viper"
)

var (
	onceConfig sync.Once
	cfg        QueueHandlersConfig
)

type QueueHandlersConfig []struct {
	Handler       string   `mapstructure:"handler"`
	WorkersCount  int      `mapstructure:"workers_count"`
	ConsumerGroup string   `mapstructure:"consumer_group"`
	DLQTopic      *string  `mapstructure:"dlq_topic"`
	Topics        []string `mapstructure:"topics"`
}

func GetConfig() QueueHandlersConfig {
	onceConfig.Do(func() {
		viper.UnmarshalKey("queue_handlers", &cfg)
	})

	if cfg == nil {
		slog.Error("common config not loaded")
	}

	return cfg
}
