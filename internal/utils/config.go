package utils

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	DatabaseDriver      string        `mapstructure:"DATABASE_DRIVER"`
	DatabaseSource      string        `mapstructure:"DATABASE_SOURCE"`
	ServerAddress       string        `mapstructure:"SERVER_ADDRESS"`
	TokenSymmetricKey   string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	AccessTokenDuration time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
}

func LoadConfig(path string, productionFlag string) (config *Config, err error) {
	viper.AddConfigPath(path)

	// detect production environment
	if productionFlag == "--production" {
		viper.SetConfigName("prod")
	} else {
		viper.SetConfigName("dev")
	}

	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)

	fmt.Println(config)

	return
}
