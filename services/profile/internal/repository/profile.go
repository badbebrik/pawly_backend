package repository

import (
	"context"
	"github.com/google/uuid"
	"profile/internal/model"
)

type ProfileRepository interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (*model.Profile, error)
	GetByUserIDs(ctx context.Context, userIDs []uuid.UUID) ([]model.Profile, error)
	Create(ctx context.Context, p *model.Profile) error
	Update(ctx context.Context, p *model.Profile) error
}
