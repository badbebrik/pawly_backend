package config

import "pawly/pkg/configenv"

type Config struct {
	AppPort string
}

func Load() *Config {
	return &Config{
		AppPort: configenv.String("APP_PORT", "8090"),
	}
}
