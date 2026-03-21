package app

import (
	"chat/internal/config"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

type App struct {
	cfg     *config.Config
	httpSrv *http.Server
}

func New(cfg *config.Config) (*App, error) {
	if cfg == nil {
		return nil, errors.New("config is nil")
	}

	a := &App{
		cfg: cfg,
	}

	a.httpSrv = &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: a.setupRoutes(),
	}

	return a, nil
}

func (a *App) Run() error {
	go func() {
		_ = a.httpSrv.ListenAndServe()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	return nil
}

func (a *App) Close() {
	if a.httpSrv != nil {
		_ = a.httpSrv.Close()
	}
}
