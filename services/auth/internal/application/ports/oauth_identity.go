package ports

import (
	"auth/internal/domain/model"
	"context"

	"github.com/google/uuid"
)

type OAuthIdentityRepository interface {
	Create(ctx context.Context, identity *model.OAuthIdentity) error
	GetByProviderAndExternalID(ctx context.Context, provider, externalID string) (*model.OAuthIdentity, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]model.OAuthIdentity, error)
	GetByEmail(ctx context.Context, provider, email string) (*model.OAuthIdentity, error)
}
