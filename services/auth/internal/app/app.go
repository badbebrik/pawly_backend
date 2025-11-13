package app

import (
	"auth/internal/config"
	"auth/internal/db"
	"github.com/rs/zerolog/log"
)

type App struct {
	Config *config.Config
	PG     *db.Postgres
}

func New(cfg *config.Config) (*App, error) {
	pg, err := db.NewPostgres(cfg)
	if err != nil {
		return nil, err
	}

	return &App{
		Config: cfg,
		PG:     pg,
	}, nil
}

func (a *App) Close() {
	log.Info().Msg("closing App resources...")
	if a.PG != nil {
		a.PG.Close()
	}
}
