package app

import (
	"context"
	"file/internal/config"
	"file/internal/infrastructure/db"
	pgrepo "file/internal/infrastructure/repository"
	"file/internal/service"
	"file/internal/storage"
	grpcserver "file/internal/transport/grpc"
	filepb "file/proto"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

type App struct {
	cfg *config.Config

	pg       *db.Postgres
	minio    *storage.MinioClient
	fileSvc  *service.FileService
	grpcSrv  *grpc.Server
	listener net.Listener
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

	listener, err := net.Listen("tcp", ":"+cfg.AppPort)
	if err != nil {
		pg.Close()
		return nil, err
	}

	grpcSrv := grpc.NewServer()
	filepb.RegisterFileServiceServer(grpcSrv, grpcserver.NewServer(fileSvc))

	app.grpcSrv = grpcSrv
	app.listener = listener

	return app, nil
}

func (a *App) Close() {
	log.Info().Msg("closing File App resources...")

	if a.grpcSrv != nil {
		a.grpcSrv.GracefulStop()
	}
	if a.listener != nil {
		_ = a.listener.Close()
	}

	if a.pg != nil {
		a.pg.Close()
	}
}

func (a *App) Run() error {
	go func() {
		log.Info().Str("port", a.cfg.AppPort).Msg("starting gRPC server")
		if err := a.grpcSrv.Serve(a.listener); err != nil {
			log.Fatal().Err(err).Msg("grpc server crash")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down File service...")

	return nil
}
