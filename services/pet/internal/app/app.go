package app

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"pet/internal/application/usecase"
	"pet/internal/config"
	aclclient "pet/internal/infrastructure/aclclient"
	"pet/internal/infrastructure/db"
	fileclient "pet/internal/infrastructure/fileclient"
	grpcserver "pet/internal/transport/grpc"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

type App struct {
	cfg *config.Config

	useCases *usecase.Set
	pg       *db.Postgres
	acl      *aclclient.Client
	file     *fileclient.Client
	httpSrv  *http.Server
	grpcSrv  *grpc.Server

	grpcListener net.Listener
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
		acl:          runtime.acl,
		file:         runtime.file,
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
	if a.grpcListener != nil {
		_ = a.grpcListener.Close()
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
		log.Info().Str("port", a.cfg.AppPort).Msg("starting pet http server")
		if err := a.httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("pet http server crash")
		}
	}()

	go func() {
		log.Info().Str("port", a.cfg.AppGRPCPort).Msg("starting pet grpc server")
		if err := a.grpcSrv.Serve(a.grpcListener); err != nil {
			log.Fatal().Err(err).Msg("pet grpc server crash")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down pet service...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.httpSrv.Shutdown(ctx); err != nil {
		return err
	}
	if a.grpcSrv != nil {
		a.grpcSrv.GracefulStop()
	}

	return nil
}
