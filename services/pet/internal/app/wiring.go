package app

import (
	"net"
	"pet/internal/application/usecase"
	"pet/internal/config"
	aclclient "pet/internal/infrastructure/aclclient"
	"pet/internal/infrastructure/db"
	fileclient "pet/internal/infrastructure/fileclient"
	pgrepo "pet/internal/infrastructure/repository"
)

type runtime struct {
	useCases     *usecase.Set
	pg           *db.Postgres
	acl          *aclclient.Client
	file         *fileclient.Client
	grpcListener net.Listener
}

func buildRuntime(cfg *config.Config) (*runtime, error) {
	pg, err := db.NewPostgres(cfg)
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

	grpcListener, err := net.Listen("tcp", ":"+cfg.AppGRPCPort)
	if err != nil {
		file.Close()
		acl.Close()
		pg.Close()
		return nil, err
	}

	return &runtime{
		useCases:     usecase.NewSet(pgrepo.NewPetRepository(pg.Pool), acl, file),
		pg:           pg,
		acl:          acl,
		file:         file,
		grpcListener: grpcListener,
	}, nil
}
