package app

import (
	"context"
	"errors"
	"file/internal/application/usecase"
	"file/internal/config"
	"file/internal/infrastructure/db"
	"file/internal/infrastructure/storage"
	"net"
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
	minio    *storage.MinioClient
	grpcSrv  *grpc.Server
	listener net.Listener
}

func New(cfg *config.Config) (*App, error) {
	rt, err := buildRuntime(cfg)
	if err != nil {
		return nil, err
	}

	return &App{
		cfg:      cfg,
		useCases: rt.useCases,
		pg:       rt.pg,
		minio:    rt.minio,
		grpcSrv:  rt.grpcSrv,
		listener: rt.listener,
	}, nil
}

func (a *App) Close() {
	log.Info().Msg("closing File App resources...")

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
		if err := a.grpcSrv.Serve(a.listener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			log.Fatal().Err(err).Msg("grpc server crash")
		}
	}()

	go a.runCleanupLoop(ctx, time.Duration(a.cfg.CleanupIntervalSeconds)*time.Second, a.cfg.CleanupBatchSize)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down File service...")
	cancel()
	if a.grpcSrv != nil {
		a.grpcSrv.GracefulStop()
	}
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
			res, err := a.useCases.RunCleanupBatch(ctx, batchSize)
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
