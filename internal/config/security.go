package config

import (
	"log/slog"
	"os"
	"sync"
)

var (
	onceSecurityConfig sync.Once
	securityCfg        *SecurityConfig
)

type SecurityConfig struct {
	SecretKey      string
	TokenSecretKey string
}

func GetSecurityConfig() *SecurityConfig {
	onceSecurityConfig.Do(func() {

		securityCfg = &SecurityConfig{
			SecretKey:      os.Getenv("SECRET_KEY"),
			TokenSecretKey: os.Getenv("TOKEN_SECRET_KEY"),
		}

	})

	if securityCfg == nil {
		slog.Error("common config not loaded")
	}

	return securityCfg
}
