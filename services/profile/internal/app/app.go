package app

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"profile/internal/application/usecase"
	"profile/internal/config"
	"profile/internal/infrastructure/db"
	"profile/internal/infrastructure/fileclient"
	grpcserver "profile/internal/transport/grpc"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

type App struct {
	cfg *config.Config

	useCases *usecase.Set
	pg       *db.Postgres
	httpSrv  *http.Server
	grpcSrv  *grpc.Server

	grpcListener net.Listener
	fileClient   *fileclient.Client
}

func New(cfg *config.Config) (*App, error) {
	runtime, err := buildRuntime(cfg)
	if err != nil {
		return nil, err
	}

	app := &App{
		cfg:          cfg,
		useCases:     runtime.useCases,
		pg:           runtime.pg,
		fileClient:   runtime.fileClient,
		grpcListener: runtime.grpcListener,
	}

	app.httpSrv = &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: app.setupRoutes(),
	}

	grpcSrv := grpc.NewServer()
	grpcserver.Register(grpcSrv, grpcserver.NewServer(runtime.useCases))
	app.grpcSrv = grpcSrv

	return app, nil
}

func (a *App) Close() {
	log.Info().Msg("closing profile app resources...")
	if a.grpcListener != nil {
		_ = a.grpcListener.Close()
	}
	if a.fileClient != nil {
		a.fileClient.Close()
	}
	if a.pg != nil {
		a.pg.Close()
	}
}

func (a *App) Run() error {
	go func() {
		log.Info().Str("port", a.cfg.AppPort).Msg("starting HTTP server")
		if err := a.httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("http server crash")
		}
	}()

	go func() {
		log.Info().Str("port", a.cfg.AppGRPCPort).Msg("starting gRPC server")
		if err := a.grpcSrv.Serve(a.grpcListener); err != nil {
			log.Fatal().Err(err).Msg("grpc server crash")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down profile service...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := a.httpSrv.Shutdown(shutdownCtx); err != nil {
		return err
	}
	if a.grpcSrv != nil {
		a.grpcSrv.GracefulStop()
	}

	return nil
}
