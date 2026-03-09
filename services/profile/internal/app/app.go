package app

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"profile/internal/config"
	"profile/internal/infrastructure/db"
	pgrepo "profile/internal/infrastructure/db/repository"
	"profile/internal/infrastructure/fileclient"
	"profile/internal/service"
	grpcserver "profile/internal/transport/grpc"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

type App struct {
	cfg *config.Config

	pg      *db.Postgres
	httpSrv *http.Server
	grpcSrv *grpc.Server

	grpcListener net.Listener
	fileClient   *fileclient.Client
}

func New(cfg *config.Config) (*App, error) {
	pg, err := db.NewPostgres(cfg)
	if err != nil {
		return nil, err
	}

	fileClient, err := fileclient.New(cfg.FileServiceGRPCAddr)
	if err != nil {
		pg.Close()
		return nil, err
	}

	profileRepo := pgrepo.NewProfileRepository(pg.Pool)
	profileSvc := service.NewService(profileRepo, cfg, fileClient)

	app := &App{
		cfg:        cfg,
		pg:         pg,
		fileClient: fileClient,
	}

	r := app.setupRoutes(profileSvc)
	app.httpSrv = &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: r,
	}

	grpcListener, err := net.Listen("tcp", ":"+cfg.AppGRPCPort)
	if err != nil {
		fileClient.Close()
		pg.Close()
		return nil, err
	}
	app.grpcListener = grpcListener

	grpcSrv := grpc.NewServer()
	grpcserver.Register(grpcSrv, grpcserver.NewServer(profileSvc))
	app.grpcSrv = grpcSrv

	return app, nil
}

func (a *App) Close() {
	log.Info().Msg("closing Profile App resources...")

	if a.httpSrv != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_ = a.httpSrv.Shutdown(ctx)
		cancel()
	}

	if a.grpcSrv != nil {
		a.grpcSrv.GracefulStop()
	}
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

	log.Info().Msg("shutting down Profile service...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := a.httpSrv.Shutdown(shutdownCtx); err != nil {
		return err
	}

	return nil
}
