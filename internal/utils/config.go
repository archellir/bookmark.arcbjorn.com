package utils

import "github.com/spf13/viper"

type Config struct {
	DatabaseDriver string `mapstructure:"DATABASE_DRIVER"`
	DatabaseSource string `mapstructure:"DATABASE_SOURCE"`
	ServerAddress  string `mapstructure:"SERVER_ADDRESS"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
