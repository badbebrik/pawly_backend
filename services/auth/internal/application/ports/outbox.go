package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type OutboxEvent struct {
	ID          uuid.UUID
	EventType   string
	Payload     []byte
	Status      string
	Attempts    int
	LastError   *string
	CreatedAt   time.Time
	PublishedAt *time.Time
}

type OutboxRepository interface {
	Create(ctx context.Context, event OutboxEvent) error
	ListPending(ctx context.Context, limit int) ([]OutboxEvent, error)
	MarkPublished(ctx context.Context, id uuid.UUID) error
	MarkFailed(ctx context.Context, id uuid.UUID, lastError string) error
}
