package config

import (
	"pawly/pkg/configenv"
	"time"
)

type Config struct {
	AppPort string

	RabbitHost       string
	RabbitPort       string
	RabbitUser       string
	RabbitPassword   string
	RabbitEmailQueue string

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

func Load() (*Config, error) {
	primaryPort, err := configenv.Int("SMTP_PRIMARY_PORT", 587)
	if err != nil {
		return nil, err
	}
	primaryUseTLS, err := configenv.Bool("SMTP_PRIMARY_USE_TLS", false)
	if err != nil {
		return nil, err
	}
	primaryUseStartTLS, err := configenv.Bool("SMTP_PRIMARY_USE_STARTTLS", true)
	if err != nil {
		return nil, err
	}
	primarySkipTLSVerify, err := configenv.Bool("SMTP_PRIMARY_SKIP_TLS_VERIFY", false)
	if err != nil {
		return nil, err
	}

	fallbackPort, err := configenv.Int("SMTP_FALLBACK_PORT", 0)
	if err != nil {
		return nil, err
	}
	fallbackUseTLS, err := configenv.Bool("SMTP_FALLBACK_USE_TLS", false)
	if err != nil {
		return nil, err
	}
	fallbackUseStartTLS, err := configenv.Bool("SMTP_FALLBACK_USE_STARTTLS", false)
	if err != nil {
		return nil, err
	}
	fallbackSkipTLSVerify, err := configenv.Bool("SMTP_FALLBACK_SKIP_TLS_VERIFY", false)
	if err != nil {
		return nil, err
	}

	requeueOnFail, err := configenv.Bool("REQUEUE_ON_FAIL", false)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		AppPort:          configenv.String("APP_PORT", ""),
		RabbitHost:       configenv.String("RABBITMQ_HOST", ""),
		RabbitPort:       configenv.String("RABBITMQ_PORT", ""),
		RabbitUser:       configenv.String("RABBITMQ_USER", ""),
		RabbitPassword:   configenv.String("RABBITMQ_PASSWORD", ""),
		RabbitEmailQueue: configenv.String("RABBITMQ_EMAIL_QUEUE", "email.notifications"),

		TemplateDir:   configenv.String("TEMPLATE_DIR", ""),
		DefaultLocale: configenv.String("DEFAULT_LOCALE", ""),

		SMTPPrimary: SMTPConfig{
			Host:           configenv.String("SMTP_PRIMARY_HOST", ""),
			Port:           primaryPort,
			Username:       configenv.String("SMTP_PRIMARY_USERNAME", ""),
			Password:       configenv.String("SMTP_PRIMARY_PASSWORD", ""),
			From:           configenv.String("SMTP_PRIMARY_FROM", ""),
			UseTLS:         primaryUseTLS,
			UseStartTLS:    primaryUseStartTLS,
			SkipTLSVerify:  primarySkipTLSVerify,
			ConnectTimeout: 5 * time.Second,
			SendTimeout:    15 * time.Second,
		},
		SMTPFallback: SMTPConfig{
			Host:           configenv.String("SMTP_FALLBACK_HOST", ""),
			Port:           fallbackPort,
			Username:       configenv.String("SMTP_FALLBACK_USERNAME", ""),
			Password:       configenv.String("SMTP_FALLBACK_PASSWORD", ""),
			From:           configenv.String("SMTP_FALLBACK_FROM", ""),
			UseTLS:         fallbackUseTLS,
			UseStartTLS:    fallbackUseStartTLS,
			SkipTLSVerify:  fallbackSkipTLSVerify,
			ConnectTimeout: 5 * time.Second,
			SendTimeout:    15 * time.Second,
		},

		RequeueOnFail: requeueOnFail,
	}

	return cfg, nil
}
