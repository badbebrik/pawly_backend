package app

import (
	"net"
	"profile/internal/application/ports"
	"profile/internal/application/usecase"
	"profile/internal/config"
	"profile/internal/infrastructure/db"
	pgrepo "profile/internal/infrastructure/db/repository"
	"profile/internal/infrastructure/fileclient"
)

type runtime struct {
	useCases     *usecase.Set
	pg           *db.Postgres
	fileClient   *fileclient.Client
	grpcListener net.Listener
}

func buildRuntime(cfg *config.Config) (*runtime, error) {
	pg, err := db.NewPostgres(cfg)
	if err != nil {
		return nil, err
	}

	fileClient, err := fileclient.New(cfg.FileServiceGRPCAddr)
	if err != nil {
		pg.Close()
		return nil, err
	}

	grpcListener, err := net.Listen("tcp", ":"+cfg.AppGRPCPort)
	if err != nil {
		fileClient.Close()
		pg.Close()
		return nil, err
	}

	return &runtime{
		useCases:     buildProfileModule(cfg, pgrepo.NewProfileRepository(pg.Pool), fileClient),
		pg:           pg,
		fileClient:   fileClient,
		grpcListener: grpcListener,
	}, nil
}

func buildProfileModule(cfg *config.Config, profiles ports.ProfileRepository, fileClient ports.FileClient) *usecase.Set {
	return usecase.NewSet(usecase.Dependencies{
		Profiles:   profiles,
		FileClient: fileClient,
		Config:     cfg,
	})
}
