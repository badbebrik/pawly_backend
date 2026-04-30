package app

import (
	"acl/internal/application/usecase"
	"acl/internal/config"
	"acl/internal/infrastructure/db"
	petclient "acl/internal/infrastructure/petclient"
	profileclient "acl/internal/infrastructure/profileclient"
	grpcserver "acl/internal/transport/grpc"
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

type App struct {
	cfg *config.Config

	useCases *usecase.Set
	pg       *db.Postgres
	profile  *profileclient.Client
	pet      *petclient.Client
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
		profile:      runtime.profile,
		pet:          runtime.pet,
		grpcListener: runtime.grpcListener,
	}

	app.httpSrv = &http.Server{
		Addr:    ":" + cfg.AppHTTPPort,
		Handler: app.setupRoutes(),
	}

	grpcSrv := grpc.NewServer()
	grpcserver.Register(grpcSrv, grpcserver.NewServer(runtime.useCases))
	app.grpcSrv = grpcSrv

	return app, nil
}

func (a *App) Close() {
	log.Info().Msg("closing acl app resources...")
	if a.grpcListener != nil {
		_ = a.grpcListener.Close()
	}

	if a.pg != nil {
		a.pg.Close()
	}
	if a.profile != nil {
		a.profile.Close()
	}
	if a.pet != nil {
		a.pet.Close()
	}
}

func (a *App) Run() error {
	go func() {
		log.Info().Str("port", a.cfg.AppHTTPPort).Msg("starting HTTP server")
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

	log.Info().Msg("shutting down acl service...")

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
