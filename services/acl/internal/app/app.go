package app

import (
	"acl/internal/config"
	"acl/internal/infrastructure/db"
	pgrepo "acl/internal/infrastructure/repository"
	aclservice "acl/internal/service"
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

	pg      *db.Postgres
	aclSvc  *aclservice.ACLService
	httpSrv *http.Server
	grpcSrv *grpc.Server

	grpcListener net.Listener
}

func New(cfg *config.Config) (*App, error) {
	pg, err := db.NewPostgres(cfg)
	if err != nil {
		return nil, err
	}

	membershipRepo := pgrepo.NewMembershipRepository(pg.Pool)
	aclSvc := aclservice.New(membershipRepo)

	app := &App{
		cfg:    cfg,
		pg:     pg,
		aclSvc: aclSvc,
	}

	httpRouter := app.setupRoutes()
	app.httpSrv = &http.Server{
		Addr:    ":" + cfg.AppHTTPPort,
		Handler: httpRouter,
	}

	grpcListener, err := net.Listen("tcp", ":"+cfg.AppGRPCPort)
	if err != nil {
		pg.Close()
		return nil, err
	}
	app.grpcListener = grpcListener

	grpcSrv := grpc.NewServer()
	grpcserver.Register(grpcSrv, grpcserver.NewServer(aclSvc))
	app.grpcSrv = grpcSrv

	return app, nil
}

func (a *App) Close() {
	log.Info().Msg("closing ACL App resources...")

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

	if a.pg != nil {
		a.pg.Close()
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

	log.Info().Msg("shutting down ACL service...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.httpSrv.Shutdown(ctx); err != nil {
		return err
	}

	return nil
}
