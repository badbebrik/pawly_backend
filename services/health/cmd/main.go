package main

import (
	"health/internal/app"
	"health/internal/config"

	"github.com/rs/zerolog/log"
)

func main() {
	cfg := config.Load()

	a, err := app.New(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("init app failed")
	}
	defer a.Close()

	if err := a.Run(); err != nil {
		log.Fatal().Err(err).Msg("app run failed")
	}
}
