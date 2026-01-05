package config

import (
	"os"
	"strconv"
	"time"
)

type SMTPConfig struct {
	Host           string
	Port           int
	Username       string
	Password       string
	From           string
	UseTLS         bool
	UseStartTLS    bool
	SkipTLSVerify  bool
	ConnectTimeout time.Duration
	SendTimeout    time.Duration
}

type RabbitConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Queue    string
}

type Config struct {
	AppPort string

	Rabbit RabbitConfig

	SMTPPrimary  SMTPConfig
	SMTPFallback SMTPConfig

	RequeueOnFail bool
	TemplateDir   string
}

func Load() *Config {
	return &Config{
		AppPort: getEnv("APP_PORT", ""),

		Rabbit: RabbitConfig{
			Host:     getEnv("RABBITMQ_HOST", ""),
			Port:     getEnv("RABBITMQ_PORT", ""),
			User:     getEnv("RABBITMQ_USER", ""),
			Password: getEnv("RABBITMQ_PASSWORD", ""),
			Queue:    getEnv("RABBITMQ_EMAIL_JOBS_QUEUE", ""),
		},

		SMTPPrimary: SMTPConfig{
			Host:           getEnv("SMTP_PRIMARY_HOST", ""),
			Port:           getEnvInt("SMTP_PRIMARY_PORT", 587),
			Username:       getEnv("SMTP_PRIMARY_USERNAME", ""),
			Password:       getEnv("SMTP_PRIMARY_PASSWORD", ""),
			From:           getEnv("SMTP_PRIMARY_FROM", ""),
			UseTLS:         getEnvBool("SMTP_PRIMARY_USE_TLS", false),
			UseStartTLS:    getEnvBool("SMTP_PRIMARY_USE_STARTTLS", true),
			SkipTLSVerify:  getEnvBool("SMTP_PRIMARY_SKIP_TLS_VERIFY", false),
			ConnectTimeout: getEnvDuration("SMTP_PRIMARY_CONNECT_TIMEOUT", 5*time.Second),
			SendTimeout:    getEnvDuration("SMTP_PRIMARY_SEND_TIMEOUT", 10*time.Second),
		},

		SMTPFallback: SMTPConfig{
			Host:           getEnv("SMTP_FALLBACK_HOST", ""),
			Port:           getEnvInt("SMTP_FALLBACK_PORT", 587),
			Username:       getEnv("SMTP_FALLBACK_USERNAME", ""),
			Password:       getEnv("SMTP_FALLBACK_PASSWORD", ""),
			From:           getEnv("SMTP_FALLBACK_FROM", ""),
			UseTLS:         getEnvBool("SMTP_FALLBACK_USE_TLS", false),
			UseStartTLS:    getEnvBool("SMTP_FALLBACK_USE_STARTTLS", true),
			SkipTLSVerify:  getEnvBool("SMTP_FALLBACK_SKIP_TLS_VERIFY", false),
			ConnectTimeout: getEnvDuration("SMTP_FALLBACK_CONNECT_TIMEOUT", 5*time.Second),
			SendTimeout:    getEnvDuration("SMTP_FALLBACK_SEND_TIMEOUT", 10*time.Second),
		},

		RequeueOnFail: getEnvBool("REQUEUE_ON_FAIL", true),
		TemplateDir:   getEnv("TEMPLATE_DIR", "./templates"),
	}
}

func getEnv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return i
}

func getEnvBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}
