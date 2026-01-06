package main

import (
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"profile/internal/app"
	"profile/internal/config"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	_ = godotenv.Load()

	cfg := config.Load()
	_ = cfg

	a, err := app.New(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to init app")
	}
	defer a.Close()

	if err := a.Run(); err != nil {
		log.Fatal().Err(err).Msg("application stopped with error")
	}
}
