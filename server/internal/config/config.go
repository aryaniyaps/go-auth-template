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

	// S3 Configuration
	S3Bucket    string `mapstructure:"S3_BUCKET"`
	S3Region    string `mapstructure:"S3_REGION"`
	S3AccessKey string `mapstructure:"S3_ACCESS_KEY"`
	S3SecretKey string `mapstructure:"S3_SECRET_KEY"`

	// SMS Configuration
	SMSProvider   string `mapstructure:"SMS_PROVIDER"`
	SMSTwilioSID  string `mapstructure:"SMS_TWILIO_SID"`
	SMSTwilioToken string `mapstructure:"SMS_TWILIO_TOKEN"`
	SMSFromNumber string `mapstructure:"SMS_FROM_NUMBER"`
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

	// Set defaults for S3 configuration
	viper.SetDefault("S3_REGION", "us-east-1")
	viper.SetDefault("SMS_PROVIDER", "dummy")

	err = viper.Unmarshal(&config)
	if err != nil {
		log.Fatal("environment can't be loaded: ", err)
	}

	return &config
}