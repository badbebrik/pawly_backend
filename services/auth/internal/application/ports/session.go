package ports

import (
	"auth/internal/domain/model"
	"context"
	"time"

	"github.com/google/uuid"
)

type SessionRepository interface {
	Create(ctx context.Context, session *model.Session) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Session, error)
	UpdateRefreshToken(ctx context.Context, id uuid.UUID, oldHash string, newHash string, newExpiresAt time.Time) error
	Revoke(ctx context.Context, id uuid.UUID) error
	RevokeAll(ctx context.Context, userID uuid.UUID) error
}
