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
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

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

	skipEnsureBucket, err := strconv.ParseBool(cfg.MinioSkipBucketEnsure)
	if err != nil {
		pg.Close()
		return nil, fmt.Errorf("parse MINIO_SKIP_BUCKET_ENSURE: %w", err)
	}

	if !skipEnsureBucket {
		if err := minioClient.EnsureBucket(context.Background()); err != nil {
			log.Error().Err(err).Msg("minio bucket init failed")
			pg.Close()
			return nil, err
		}
		log.Info().Str("bucket", minioClient.Bucket).Msg("minio bucket ready")
	} else {
		log.Warn().Str("bucket", minioClient.Bucket).Msg("minio bucket ensure is disabled by config")
	}

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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		log.Info().Str("port", a.cfg.AppPort).Msg("starting gRPC server")
		if err := a.grpcSrv.Serve(a.listener); err != nil {
			log.Fatal().Err(err).Msg("grpc server crash")
		}
	}()

	cleanupInterval, err := strconv.Atoi(a.cfg.CleanupIntervalSeconds)
	if err != nil {
		return fmt.Errorf("parse CLEANUP_INTERVAL_SECONDS: %w", err)
	}
	if cleanupInterval <= 0 {
		return fmt.Errorf("parse CLEANUP_INTERVAL_SECONDS: must be > 0")
	}
	cleanupBatchSize, err := strconv.Atoi(a.cfg.CleanupBatchSize)
	if err != nil {
		return fmt.Errorf("parse CLEANUP_BATCH_SIZE: %w", err)
	}
	if cleanupBatchSize <= 0 {
		return fmt.Errorf("parse CLEANUP_BATCH_SIZE: must be > 0")
	}

	go a.runCleanupLoop(ctx, time.Duration(cleanupInterval)*time.Second, cleanupBatchSize)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down File service...")
	cancel()

	return nil
}

func (a *App) runCleanupLoop(ctx context.Context, interval time.Duration, batchSize int) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			res, err := a.fileSvc.RunCleanupBatch(ctx, batchSize)
			if err != nil {
				log.Error().Err(err).Msg("file cleanup batch failed")
				continue
			}
			if res.DeletedPending == 0 && res.MarkedPendingDelete == 0 && res.DeletedExpired == 0 {
				continue
			}
			log.Info().
				Int("deleted_pending", res.DeletedPending).
				Int("still_pending", res.MarkedPendingDelete).
				Int("deleted_expired", res.DeletedExpired).
				Msg("file cleanup batch completed")
		}
	}
}
