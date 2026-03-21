package jwt_manager

import (
	"log/slog"
	"sync"

	"github.com/spf13/viper"
)

var (
	onceAuthConfig sync.Once
	authCfg        *AuthConfig
)

type AuthConfig struct {
	RefreshTokenTtlDays   int
	AccessTokenTtlMinutes int
}

func GetConfig() *AuthConfig {
	onceAuthConfig.Do(func() {

		AccessTokenTtlMinutes := viper.GetInt("auth.access_token_ttl_minutes")
		if AccessTokenTtlMinutes == 0 {
			AccessTokenTtlMinutes = 12
		}

		refreshTokenTtlDays := viper.GetInt("auth.refresh_token_ttl_days")
		if refreshTokenTtlDays == 0 {
			refreshTokenTtlDays = 30
		}

		authCfg = &AuthConfig{
			AccessTokenTtlMinutes: AccessTokenTtlMinutes,
			RefreshTokenTtlDays:   refreshTokenTtlDays,
		}
	})

	if authCfg == nil {
		slog.Error("common config not loaded")
	}

	return authCfg
}
