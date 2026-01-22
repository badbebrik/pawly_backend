package app

import (
	"context"
	"errors"
	"file/internal/config"
	"file/internal/infrastructure/db"
	pgrepo "file/internal/infrastructure/repository"
	"file/internal/service"
	"file/internal/storage"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
)

type App struct {
	cfg *config.Config

	pg      *db.Postgres
	httpSrv *http.Server
	minio   *storage.MinioClient
	fileSvc *service.FileService
}

func New(cfg *config.Config) (*App, error) {
	pg, err := db.NewPostgres(cfg)
	if err != nil {
		return nil, err
	}

	minioClient, err := storage.NewMinio(cfg)
	if err != nil {
		pg.Close()
		return nil, err
	}

	storageAdapter := storage.NewMinioStorageAdapter(minioClient)

	fileObjRepo := pgrepo.NewFileObjectRepository(pg.Pool)
	fileLinkRepo := pgrepo.NewFileLinkRepository(pg.Pool)

	fileSvc := service.NewFileService(fileObjRepo, fileLinkRepo, storageAdapter)

	if err := minioClient.EnsureBucket(context.Background()); err != nil {
		log.Error().Err(err).Msg("minio bucket init failed")
		pg.Close()
		return nil, err
	}
	log.Info().Str("bucket", minioClient.Bucket).Msg("minio bucket ready")

	app := &App{
		cfg:     cfg,
		pg:      pg,
		minio:   minioClient,
		fileSvc: fileSvc,
	}

	r := app.setupRoutes()
	app.httpSrv = &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: r,
	}

	return app, nil
}

func (a *App) Close() {
	log.Info().Msg("closing File App resources...")

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
	go func() {
		log.Info().Str("port", a.cfg.AppPort).Msg("starting HTTP server")
		if err := a.httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("http server crash")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down File service...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := a.httpSrv.Shutdown(shutdownCtx); err != nil {
		return err
	}

	return nil
}
