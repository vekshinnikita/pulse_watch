package redis_db

import (
	"log/slog"
	"os"
	"sync"

	"github.com/spf13/viper"
)

type RedisConfig struct {
	Host     string
	Port     int
	Password string
}

var (
	cfg  *RedisConfig
	once sync.Once
)

func GetConfig() *RedisConfig {
	once.Do(func() {
		host := viper.GetString("redis.host")
		if host == "" {
			host = "localhost"
		}

		port := viper.GetInt("redis.port")
		if port == 0 {
			port = 6379
		}

		cfg = &RedisConfig{
			Host:     host,
			Port:     port,
			Password: os.Getenv("REDIS_PASSWORD"),
		}
	})

	if cfg == nil {
		slog.Error("server config not loaded")
	}

	return cfg
}
