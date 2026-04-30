package app

import (
	"acl/internal/application/usecase"
	"acl/internal/config"
	"acl/internal/infrastructure/db"
	petclient "acl/internal/infrastructure/petclient"
	profileclient "acl/internal/infrastructure/profileclient"
	pgrepo "acl/internal/infrastructure/repository"
	"net"
	"time"
)

type runtime struct {
	useCases     *usecase.Set
	pg           *db.Postgres
	profile      *profileclient.Client
	pet          *petclient.Client
	grpcListener net.Listener
}

func buildRuntime(cfg *config.Config) (*runtime, error) {
	pg, err := db.NewPostgres(cfg)
	if err != nil {
		return nil, err
	}

	profile, err := profileclient.New(cfg.ProfileServiceGRPCAddr)
	if err != nil {
		pg.Close()
		return nil, err
	}

	pet, err := petclient.New(cfg.PetServiceGRPCAddr)
	if err != nil {
		profile.Close()
		pg.Close()
		return nil, err
	}

	grpcListener, err := net.Listen("tcp", ":"+cfg.AppGRPCPort)
	if err != nil {
		pet.Close()
		profile.Close()
		pg.Close()
		return nil, err
	}

	useCases := usecase.NewSet(usecase.Dependencies{
		Memberships: pgrepo.NewMembershipRepository(pg.Pool),
		Roles:       pgrepo.NewRoleRepository(pg.Pool),
		Invites:     pgrepo.NewInviteRepository(pg.Pool),
		Options: usecase.Options{
			InviteTTL:          time.Duration(cfg.InviteTTLMinutes) * time.Minute,
			InviteDeeplinkBase: cfg.InviteDeeplinkBase,
		},
	})

	return &runtime{
		useCases:     useCases,
		pg:           pg,
		profile:      profile,
		pet:          pet,
		grpcListener: grpcListener,
	}, nil
}
