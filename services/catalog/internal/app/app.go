package app

import (
	"catalog/infrastructure/db"
	"catalog/internal/config"
	"context"
	"errors"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type App struct {
	cfg *config.Config
	pg  *db.Postgres

	httpSrv *http.Server
}

func New(cfg *config.Config) (*App, error) {
	pg, err := db.NewPostgres(cfg)
	if err != nil {
		return nil, err
	}

	return &App{
		cfg: cfg,
		pg:  pg,
	}, nil
}

func (a *App) Close() {
	log.Info().Msg("closing Catalog App resources...")

	if a.httpSrv != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_ = a.httpSrv.Shutdown(ctx)
		cancel()
	}

	if a.pg != nil {
		a.pg.Close()
	}
}

func (a *App) Run() error {
	a.httpSrv = &http.Server{
		Addr:    ":" + a.cfg.AppPort,
		Handler: a.routes(),
	}

	go func() {
		log.Info().Str("port", a.cfg.AppPort).Msg("starting HTTP server")
		if err := a.httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("http server crash")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down Catalog service...")
	return nil
}
