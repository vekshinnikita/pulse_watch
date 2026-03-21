package postgres

import (
	"log/slog"
	"os"
	"sync"

	"github.com/spf13/viper"
)

type PostgresConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	DBName   string
	SSLMode  string
}

var (
	cfg  *PostgresConfig
	once sync.Once
)

func GetConfig() *PostgresConfig {
	once.Do(func() {
		host := viper.GetString("database.host")
		if host == "" {
			host = "localhost"
		}

		port := viper.GetInt("database.port")
		if port == 0 {
			port = 5432
		}

		sslMode := viper.GetBool("database.ssl_mode")
		sslModeString := "disable"
		if sslMode {
			sslModeString = "enable"
		}

		cfg = &PostgresConfig{
			Host:     host,
			Port:     port,
			Username: viper.GetString("database.username"),
			Password: os.Getenv("DB_PASSWORD"),
			DBName:   viper.GetString("database.db_name"),
			SSLMode:  sslModeString,
		}
	})

	if cfg == nil {
		slog.Error("server config not loaded")
	}

	return cfg
}
