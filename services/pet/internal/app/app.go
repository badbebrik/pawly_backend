package app

import (
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"pet/internal/config"
	"pet/internal/infrastructure"
	aclclient "pet/internal/infrastructure/aclclient"
	fileclient "pet/internal/infrastructure/fileclient"
	pgrepo "pet/internal/infrastructure/repository"
	"pet/internal/service"
	grpcserver "pet/internal/transport/grpc"
	"syscall"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

type App struct {
	cfg *config.Config

	pg   *infrastructure.Postgres
	acl  *aclclient.Client
	file *fileclient.Client

	petSvc       *service.PetService
	httpSrv      *http.Server
	grpcSrv      *grpc.Server
	grpcListener net.Listener
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

	petRepo := pgrepo.NewPetRepository(pg.Pool)
	petSvc := service.New(petRepo, acl, file)

	app := &App{
		cfg:    cfg,
		pg:     pg,
		acl:    acl,
		file:   file,
		petSvc: petSvc,
	}

	app.httpSrv = &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: app.setupRoutes(),
	}

	grpcListener, err := net.Listen("tcp", ":"+cfg.AppGRPCPort)
	if err != nil {
		file.Close()
		acl.Close()
		pg.Close()
		return nil, err
	}
	app.grpcListener = grpcListener

	grpcSrv := grpc.NewServer()
	grpcserver.Register(grpcSrv, grpcserver.NewServer(petSvc))
	app.grpcSrv = grpcSrv

	return app, nil
}

func (a *App) Close() {
	if a.httpSrv != nil {
		_ = a.httpSrv.Close()
	}
	if a.grpcSrv != nil {
		a.grpcSrv.GracefulStop()
	}
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
		log.Info().Str("port", a.cfg.AppPort).Msg("starting Pet HTTP server")
		if err := a.httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("pet http server crash")
		}
	}()

	go func() {
		log.Info().Str("port", a.cfg.AppGRPCPort).Msg("starting Pet gRPC server")
		if err := a.grpcSrv.Serve(a.grpcListener); err != nil {
			log.Fatal().Err(err).Msg("pet grpc server crash")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down Pet service...")
	return nil
}
