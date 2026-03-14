package usecase

import (
	"profile/internal/application/ports"
	"profile/internal/config"
)

type Dependencies struct {
	Profiles ports.ProfileRepository
	Files    ports.FileGateway
	Config   *config.Config
}

type dependencies struct {
	profiles ports.ProfileRepository
	files    ports.FileGateway
	config   *config.Config
}

func newDependencies(in Dependencies) *dependencies {
	return &dependencies{
		profiles: in.Profiles,
		files:    in.Files,
		config:   in.Config,
	}
}
