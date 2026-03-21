package es

import (
	"log/slog"
	"os"
	"sync"

	"github.com/spf13/viper"
)

type ElasticSearchConfig struct {
	Addresses []string
	APIKey    string
}

var (
	cfg  *ElasticSearchConfig
	once sync.Once
)

func GetConfig() *ElasticSearchConfig {
	once.Do(func() {
		cfg = &ElasticSearchConfig{
			Addresses: viper.GetStringSlice("elasticsearch.addresses"),
			APIKey:    os.Getenv("ELASTIC_API_KEY"),
		}
	})

	if cfg == nil {
		slog.Error("server config not loaded")
	}

	return cfg
}
