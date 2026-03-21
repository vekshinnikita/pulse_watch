package config

import (
	"fmt"
	"sync"

	"github.com/spf13/viper"
)

var (
	onceLoadConfig sync.Once
)

func LoadConfig(path string) error {
	var err error
	onceLoadConfig.Do(func() {
		viper.SetConfigFile(path)
		viper.SetConfigType("yaml")

		if readErr := viper.ReadInConfig(); readErr != nil {
			err = fmt.Errorf("error reading config file: %w", readErr)
			return
		}
	})

	return err
}
