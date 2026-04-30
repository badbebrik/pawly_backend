package main

import (
	"email/internal/app"
	"email/internal/config"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}
	a, err := app.New(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to init app")
	}
	defer a.Close()

	if err := a.Run(); err != nil {
		log.Fatal().Err(err).Msg("application stopped with error")
	}
}
