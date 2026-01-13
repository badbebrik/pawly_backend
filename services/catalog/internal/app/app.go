package app

import (
	"catalog/internal/config"
	"catalog/internal/infrastructure/db"
	dbrepo "catalog/internal/infrastructure/db/repository"
	"catalog/internal/service"
	"catalog/internal/transport/http/handlers"
	"context"
	"errors"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type App struct {
	cfg *config.Config
	pg  *db.Postgres

	CatalogSvc *service.CatalogService

	PublicHandler *handlers.PublicHandler
	AdminSpecies  *handlers.AdminSpeciesHandler
	AdminColors   *handlers.AdminColorHandler
	AdminPatterns *handlers.AdminPatternHandler

	httpSrv *http.Server
}

func New(cfg *config.Config) (*App, error) {
	pg, err := db.NewPostgres(cfg)
	if err != nil {
		return nil, err
	}

	versionRepo := dbrepo.NewVersionRepo(pg.Pool)
	speciesRepo := dbrepo.NewSpeciesRepo(pg.Pool)
	colorRepo := dbrepo.NewColorRepo(pg.Pool)
	patternRepo := dbrepo.NewPatternRepo(pg.Pool)

	catalogSvc := service.NewCatalogService(pg.Pool, versionRepo, speciesRepo, colorRepo, patternRepo)

	publicHandler := handlers.NewPublicHandler(catalogSvc)
	adminSpeciesHandler := handlers.NewAdminSpeciesHandler(catalogSvc, speciesRepo)
	adminColorHandler := handlers.NewAdminColorHandler(catalogSvc, colorRepo)
	adminPatternHandler := handlers.NewAdminPatternHandler(catalogSvc, patternRepo)

	return &App{
		cfg:           cfg,
		pg:            pg,
		CatalogSvc:    catalogSvc,
		PublicHandler: publicHandler,
		AdminSpecies:  adminSpeciesHandler,
		AdminColors:   adminColorHandler,
		AdminPatterns: adminPatternHandler,
	}, nil
}

func (a *App) Close() {
	log.Info().Msg("closing Catalog App resources...")

	if a.httpSrv != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_ = a.httpSrv.Shutdown(ctx)
		cancel()
	}

	if a.pg != nil {
		a.pg.Close()
	}
}

func (a *App) Run() error {
	a.httpSrv = &http.Server{
		Addr:    ":" + a.cfg.AppPort,
		Handler: a.routes(),
	}

	go func() {
		log.Info().Str("port", a.cfg.AppPort).Msg("starting HTTP server")
		if err := a.httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("http server crash")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down Catalog service...")
	return nil
}
