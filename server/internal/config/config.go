package config

import (
	"log"
	"path/filepath"
	"runtime"

	"github.com/spf13/viper"
)

type Config struct {
	ServerPort  string `mapstructure:"SERVER_PORT"`
	Environment string `mapstructure:"ENVIRONMENT"`

	DBUrl string `mapstructure:"DB_URL"`
}

func SetupConfig() *Config {
	config := Config{}
	_, b, _, _ := runtime.Caller(0)
	basePath := filepath.Join(filepath.Dir(b), "../../")
	viper.SetConfigFile(filepath.Join(basePath, ".env"))

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal("cannot read configuration", err)
	}

	viper.SetDefault("TIMEZONE", "UTC")

	err = viper.Unmarshal(&config)
	if err != nil {
		log.Fatal("environment can't be loaded: ", err)
	}

	return &config
}
