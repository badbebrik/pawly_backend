package app

import (
	"context"
	"errors"
	"gateway/internal/config"
	"gateway/internal/grpc"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type App struct {
	cfg *config.Config

	httpSrv *http.Server

	fileClient *grpc.FileClient
}

func New(cfg *config.Config) (*App, error) {
	fileClient, err := grpc.NewFileClient(cfg.FileServiceGRPCAddr)
	if err != nil {
		return nil, err
	}

	app := &App{
		cfg:        cfg,
		fileClient: fileClient,
	}

	r := app.setupRoutes()
	app.httpSrv = &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: r,
	}

	return app, nil
}

func (a *App) Close() {
	log.Info().Msg("closing Gateway resources...")

	if a.httpSrv != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_ = a.httpSrv.Shutdown(ctx)
		cancel()
	}
	if a.fileClient != nil {
		a.fileClient.Close()
	}
}

func (a *App) Run() error {
	go func() {
		log.Info().Str("port", a.cfg.AppPort).Msg("starting HTTP gateway")
		if err := a.httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("http server crash")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down Gateway...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	_ = shutdownCtx

	return nil
}
