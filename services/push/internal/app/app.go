package app

import (
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"push/internal/config"
	"push/internal/infrastructure"

	"github.com/rs/zerolog/log"
)

type App struct {
	cfg *config.Config
	pg  *infrastructure.Postgres

	httpSrv *http.Server
}

func New(cfg *config.Config) (*App, error) {
	pg, err := infrastructure.NewPostgres(cfg)
	if err != nil {
		return nil, err
	}

	a := &App{
		cfg: cfg,
		pg:  pg,
	}
	a.httpSrv = &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: a.setupRoutes(),
	}
	return a, nil
}

func (a *App) Close() {
	if a.httpSrv != nil {
		_ = a.httpSrv.Close()
	}
	if a.pg != nil {
		a.pg.Close()
	}
}

func (a *App) Run() error {
	go func() {
		log.Info().Str("port", a.cfg.AppPort).Msg("starting Push HTTP server")
		if err := a.httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("push http server crash")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down Push service...")
	return nil
}
