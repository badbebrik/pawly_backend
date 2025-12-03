package auth

import (
	"auth/internal/notifications"
	"auth/internal/repository/device_repo"
	"auth/internal/repository/oauth_identity_repo"
	"auth/internal/repository/session_repo"
	"auth/internal/repository/user_repo"
	"auth/internal/tokens"
	"auth/internal/verification"
	"context"
	"github.com/google/uuid"
)

type Service struct {
	users        user_repo.Repository
	sessions     session_repo.Repository
	oauth        oauth_identity_repo.Repository
	devices      device_repo.Repository
	verification verification.Repository
	notifier     notifications.Publisher
	jwt          *tokens.JWTService
	profiles     ProfileClient
}

type ProfileClient interface {
	CreateProfile(ctx context.Context, userID uuid.UUID, firstName, lastName string) error
}

func NewService(
	users user_repo.Repository,
	sessions session_repo.Repository,
	oauth oauth_identity_repo.Repository,
	devices device_repo.Repository,
	verification verification.Repository,
	notifier notifications.Publisher,
	jwt *tokens.JWTService,
	profiles ProfileClient,
) *Service {
	return &Service{
		users:        users,
		sessions:     sessions,
		oauth:        oauth,
		devices:      devices,
		verification: verification,
		notifier:     notifier,
		jwt:          jwt,
		profiles:     profiles,
	}
}
