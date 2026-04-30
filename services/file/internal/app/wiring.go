package app

import (
	"context"
	"file/internal/application/usecase"
	"file/internal/config"
	"file/internal/infrastructure/db"
	pgrepo "file/internal/infrastructure/repository"
	"file/internal/infrastructure/storage"
	grpcserver "file/internal/transport/grpc"
	"net"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	filepb "pawly/pkg/filepb"
)

type runtime struct {
	useCases *usecase.Set
	pg       *db.Postgres
	minio    *storage.MinioClient
	grpcSrv  *grpc.Server
	listener net.Listener
}

func buildRuntime(cfg *config.Config) (*runtime, error) {
	pg, err := db.NewPostgres(cfg)
	if err != nil {
		return nil, err
	}

	minioClient, err := storage.NewMinio(cfg)
	if err != nil {
		pg.Close()
		return nil, err
	}

	if !cfg.MinioSkipBucketEnsure {
		if err := minioClient.EnsureBucket(context.Background()); err != nil {
			pg.Close()
			return nil, err
		}
		log.Info().Str("bucket", minioClient.Bucket).Msg("minio bucket ready")
	} else {
		log.Warn().Str("bucket", minioClient.Bucket).Msg("minio bucket ensure is disabled by config")
	}

	useCases := usecase.NewSet(usecase.Dependencies{
		Objects: pgrepo.NewFileObjectRepository(pg.Pool),
		Links:   pgrepo.NewFileLinkRepository(pg.Pool),
		Storage: storage.NewMinioStorageAdapter(minioClient),
	})

	listener, err := net.Listen("tcp", ":"+cfg.AppPort)
	if err != nil {
		pg.Close()
		return nil, err
	}

	grpcSrv := grpc.NewServer()
	filepb.RegisterFileServiceServer(grpcSrv, grpcserver.NewServer(useCases))

	return &runtime{
		useCases: useCases,
		pg:       pg,
		minio:    minioClient,
		grpcSrv:  grpcSrv,
		listener: listener,
	}, nil
}
