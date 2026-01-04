package service

import (
	"auth/internal/infrastructure/rabbit"
	"auth/internal/infrastructure/tokens"
	"auth/internal/repository"
	"auth/internal/verification"
)

type Service struct {
	users        repository.UserRepository
	sessions     repository.SessionRepository
	oauth        repository.OAuthIdentityRepository
	devices      repository.DeviceRepository
	verification verification.Store
	notifier     rabbit.Publisher
	jwt          tokens.TokenManager
}

func NewService(
	users repository.UserRepository,
	sessions repository.SessionRepository,
	oauth repository.OAuthIdentityRepository,
	devices repository.DeviceRepository,
	verification verification.Store,
	notifier rabbit.Publisher,
	jwt tokens.TokenManager,
) *Service {
	return &Service{
		users:        users,
		sessions:     sessions,
		oauth:        oauth,
		devices:      devices,
		verification: verification,
		notifier:     notifier,
		jwt:          jwt,
	}
}
