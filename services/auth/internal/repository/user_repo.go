package repository

import (
	"auth/internal/model"
	"context"
	"github.com/google/uuid"
	"time"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	SetVerified(ctx context.Context, id uuid.UUID) error
	UpdatePasswordHash(ctx context.Context, id uuid.UUID, newHash string) error
	UpdateEmail(ctx context.Context, id uuid.UUID, newEmail string) error
	SetActive(ctx context.Context, id uuid.UUID, active bool) error
	UpdateLastLoginAt(ctx context.Context, id uuid.UUID, t time.Time) error
}
