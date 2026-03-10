package app

import (
	"errors"
	"health/internal/config"
	"health/internal/infrastructure"
	"health/internal/infrastructure/aclclient"
	"health/internal/infrastructure/fileclient"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"
)

type App struct {
	cfg *config.Config

	pg   *infrastructure.Postgres
	acl  *aclclient.Client
	file *fileclient.Client

	httpSrv *http.Server
}

func New(cfg *config.Config) (*App, error) {
	pg, err := infrastructure.NewPostgres(cfg)
	if err != nil {
		return nil, err
	}

	acl, err := aclclient.New(cfg.ACLGRPCAddr)
	if err != nil {
		pg.Close()
		return nil, err
	}

	file, err := fileclient.New(cfg.FileGRPCAddr)
	if err != nil {
		acl.Close()
		pg.Close()
		return nil, err
	}

	a := &App{
		cfg:  cfg,
		pg:   pg,
		acl:  acl,
		file: file,
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
	if a.acl != nil {
		a.acl.Close()
	}
	if a.file != nil {
		a.file.Close()
	}
	if a.pg != nil {
		a.pg.Close()
	}
}

func (a *App) Run() error {
	go func() {
		log.Info().Str("port", a.cfg.AppPort).Msg("starting Health HTTP server")
		if err := a.httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("health http server crash")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down Health service...")
	return nil
}
