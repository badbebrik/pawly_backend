package app

import (
	"auth/internal/config"
	"auth/internal/db"
	"github.com/rs/zerolog/log"
)

type App struct {
	Config *config.Config
	PG     *db.Postgres
	Redis  *db.Redis
}

func New(cfg *config.Config) (*App, error) {
	pg, err := db.NewPostgres(cfg)
	if err != nil {
		return nil, err
	}

	redis, err := db.NewRedis(cfg)
	if err != nil {
		return nil, err
	}

	return &App{
		Config: cfg,
		PG:     pg,
		Redis:  redis,
	}, nil
}

func (a *App) Close() {
	log.Info().Msg("closing App resources...")

	if a.Redis != nil {
		err := a.Redis.Close()
		if err != nil {
			return
		}
	}

	if a.PG != nil {
		a.PG.Close()
	}
}
