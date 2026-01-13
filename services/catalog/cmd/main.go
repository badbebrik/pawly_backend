package main

import (
	"catalog/internal/app"
	"catalog/internal/config"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	_ = godotenv.Load()

	cfg := config.Load()

	a, err := app.New(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to init app")
	}
	defer a.Close()

	if err := a.Run(); err != nil {
		log.Fatal().Err(err).Msg("application stopped with error")
	}
}
