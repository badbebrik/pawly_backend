package usecase

import (
	"profile/internal/application/ports"
	"profile/internal/config"
)

type Dependencies struct {
	Profiles   ports.ProfileRepository
	FileClient ports.FileClient
	Config     *config.Config
}

type dependencies struct {
	profiles   ports.ProfileRepository
	fileClient ports.FileClient
	config     *config.Config
}

func newDependencies(in Dependencies) *dependencies {
	return &dependencies{
		profiles:   in.Profiles,
		fileClient: in.FileClient,
		config:     in.Config,
	}
}
