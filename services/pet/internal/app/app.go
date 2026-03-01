package app

import (
	"errors"
	"net/http"
	"os"
	"os/signal"
	"pet/internal/config"
	"pet/internal/infrastructure"
	aclclient "pet/internal/infrastructure/aclclient"
	fileclient "pet/internal/infrastructure/fileclient"
	pgrepo "pet/internal/infrastructure/repository"
	"pet/internal/service"
	"syscall"

	"github.com/rs/zerolog/log"
)

type App struct {
	cfg *config.Config

	pg   *infrastructure.Postgres
	acl  *aclclient.Client
	file *fileclient.Client

	petSvc  *service.PetService
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

	return app, nil
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
		log.Info().Str("port", a.cfg.AppPort).Msg("starting Pet HTTP server")
		if err := a.httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("pet http server crash")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down Pet service...")
	return nil
}
