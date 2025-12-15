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

	// JWT Configuration
	JWTSecret string `mapstructure:"JWT_SECRET"`

	// JWE Configuration
	JWESecret string `mapstructure:"JWE_SECRET"`

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

	// Email Configuration
	EmailProvider     string `mapstructure:"EMAIL_PROVIDER"`
	SMTPHost          string `mapstructure:"SMTP_HOST"`
	SMTPPort          int    `mapstructure:"SMTP_PORT"`
	SMTPUsername      string `mapstructure:"SMTP_USERNAME"`
	SMTPPassword      string `mapstructure:"SMTP_PASSWORD"`
	FromEmail         string `mapstructure:"FROM_EMAIL"`
	FromName          string `mapstructure:"FROM_NAME"`
	EmailTemplatePath string `mapstructure:"EMAIL_TEMPLATE_PATH"`

	// SES Configuration
	SESRegion         string `mapstructure:"SES_REGION"`
	SESAccessKeyID    string `mapstructure:"SES_ACCESS_KEY_ID"`
	SESSecretAccessKey string `mapstructure:"SES_SECRET_ACCESS_KEY"`

	// SendGrid Configuration
	SendGridAPIKey string `mapstructure:"SENDGRID_API_KEY"`

	// Captcha Configuration
	CaptchaProvider       string `mapstructure:"CAPTCHA_PROVIDER"`
	CloudflareSecretKey   string `mapstructure:"CLOUDFLARE_SECRET_KEY"`
	CloudflareSiteKey     string `mapstructure:"CLOUDFLARE_SITEKEY"`
	HCaptchaSecretKey     string `mapstructure:"HCAPTCHA_SECRET_KEY"`
	HCaptchaSiteKey       string `mapstructure:"HCAPTCHA_SITEKEY"`
	ReCaptchaSecretKey    string `mapstructure:"RECAPTCHA_SECRET_KEY"`
	ReCaptchaSiteKey      string `mapstructure:"RECAPTCHA_SITEKEY"`
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

	// Set default for JWT secret in development
	viper.SetDefault("JWT_SECRET", "development-secret-change-in-production")

	// Set default for JWE secret in development (exactly 16 bytes for AES-128)
	viper.SetDefault("JWE_SECRET", "dev-jwe-secret-16b!")

	// Set defaults for email configuration
	viper.SetDefault("EMAIL_PROVIDER", "dummy")
	viper.SetDefault("EMAIL_TEMPLATE_PATH", "./templates/emails")
	viper.SetDefault("SMTP_PORT", 587)

	// Set defaults for captcha configuration
	viper.SetDefault("CAPTCHA_PROVIDER", "dummy")

	err = viper.Unmarshal(&config)
	if err != nil {
		log.Fatal("environment can't be loaded: ", err)
	}

	return &config
}