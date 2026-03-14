package ports

import (
	"context"

	"github.com/google/uuid"
)

type ProfileProvisioner interface {
	CreateProfile(ctx context.Context, userID uuid.UUID, locale string, firstName string, lastName string) error
	DeleteProfile(ctx context.Context, userID uuid.UUID) error
}
