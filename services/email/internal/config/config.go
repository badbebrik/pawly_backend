package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppPort string

	RabbitHost           string
	RabbitPort           string
	RabbitUser           string
	RabbitPassword       string
	RabbitEmailJobsQueue string

	TemplateDir   string
	DefaultLocale string

	SMTPPrimary  SMTPConfig
	SMTPFallback SMTPConfig

	RequeueOnFail bool
}

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

func Load() *Config {
	return &Config{
		AppPort:              getEnv("APP_PORT", ""),
		RabbitHost:           getEnv("RABBITMQ_HOST", ""),
		RabbitPort:           getEnv("RABBITMQ_PORT", ""),
		RabbitUser:           getEnv("RABBITMQ_USER", ""),
		RabbitPassword:       getEnv("RABBITMQ_PASSWORD", ""),
		RabbitEmailJobsQueue: getEnv("RABBITMQ_EMAIL_JOBS_QUEUE", ""),

		TemplateDir:   getEnv("TEMPLATE_DIR", ""),
		DefaultLocale: getEnv("DEFAULT_LOCALE", ""),

		SMTPPrimary: SMTPConfig{
			Host:           getEnv("SMTP_PRIMARY_HOST", ""),
			Port:           getEnvInt("SMTP_PRIMARY_PORT", 587),
			Username:       getEnv("SMTP_PRIMARY_USERNAME", ""),
			Password:       getEnv("SMTP_PRIMARY_PASSWORD", ""),
			From:           getEnv("SMTP_PRIMARY_FROM", ""),
			UseTLS:         getEnvBool("SMTP_PRIMARY_USE_TLS", false),
			UseStartTLS:    getEnvBool("SMTP_PRIMARY_USE_STARTTLS", true),
			SkipTLSVerify:  getEnvBool("SMTP_PRIMARY_SKIP_TLS_VERIFY", false),
			ConnectTimeout: 5 * time.Second,
			SendTimeout:    15 * time.Second,
		},
		SMTPFallback: SMTPConfig{
			Host:           getEnv("SMTP_FALLBACK_HOST", ""),
			Port:           getEnvInt("SMTP_FALLBACK_PORT", 0),
			Username:       getEnv("SMTP_FALLBACK_USERNAME", ""),
			Password:       getEnv("SMTP_FALLBACK_PASSWORD", ""),
			From:           getEnv("SMTP_FALLBACK_FROM", ""),
			UseTLS:         getEnvBool("SMTP_FALLBACK_USE_TLS", false),
			UseStartTLS:    getEnvBool("SMTP_FALLBACK_USE_STARTTLS", false),
			SkipTLSVerify:  getEnvBool("SMTP_FALLBACK_SKIP_TLS_VERIFY", false),
			ConnectTimeout: 5 * time.Second,
			SendTimeout:    15 * time.Second,
		},

		RequeueOnFail: getEnvBool("REQUEUE_ON_FAIL", false),
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v := getEnv(key, "")
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
	v := getEnv(key, "")
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}
