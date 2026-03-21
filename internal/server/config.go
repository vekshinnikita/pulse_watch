package server

import (
	"log/slog"
	"sync"

	"github.com/spf13/viper"
)

type ServerConfig struct {
	Host string
	Port int
}

var (
	webCfg *ServerConfig
	wsCfg  *ServerConfig
	once   sync.Once
)

func GetWebConfig() *ServerConfig {
	once.Do(func() {
		host := viper.GetString("server.host")
		if host == "" {
			host = "localhost" // дефолт
		}

		port := viper.GetInt("server.port")
		if port == 0 {
			port = 8000 // дефолт
		}

		webCfg = &ServerConfig{
			Host: host,
			Port: port,
		}
	})

	if webCfg == nil {
		slog.Error("server config not loaded")
	}

	return webCfg
}

func GetWSConfig() *ServerConfig {
	once.Do(func() {
		host := viper.GetString("ws.host")
		if host == "" {
			host = "localhost" // дефолт
		}

		port := viper.GetInt("ws.port")
		if port == 0 {
			port = 8001 // дефолт
		}

		wsCfg = &ServerConfig{
			Host: host,
			Port: port,
		}
	})

	if wsCfg == nil {
		slog.Error("server config not loaded")
	}

	return wsCfg
}
