package pgrepo

import (
	"auth/internal/repository"
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OutboxRepo struct {
	db *pgxpool.Pool
}

func NewOutboxRepo(db *pgxpool.Pool) *OutboxRepo {
	return &OutboxRepo{db: db}
}

func (r *OutboxRepo) Create(ctx context.Context, event repository.OutboxEvent) error {
	const query = `
        INSERT INTO outbox_events (
            id, event_type, payload, status, attempts, last_error, created_at, published_at
        ) VALUES ($1, $2, $3, $4, $5, $6, NOW(), NULL)
    `

	_, err := r.db.Exec(ctx, query,
		event.ID,
		event.EventType,
		event.Payload,
		"PENDING",
		0,
		nil,
	)
	return err
}

func (r *OutboxRepo) ListPending(ctx context.Context, limit int) ([]repository.OutboxEvent, error) {
	if limit <= 0 {
		limit = 100
	}

	const query = `
        SELECT id, event_type, payload, status, attempts, last_error, created_at, published_at
        FROM outbox_events
        WHERE published_at IS NULL
        ORDER BY created_at ASC
        LIMIT $1
    `

	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]repository.OutboxEvent, 0, limit)
	for rows.Next() {
		var item repository.OutboxEvent
		if err := rows.Scan(
			&item.ID,
			&item.EventType,
			&item.Payload,
			&item.Status,
			&item.Attempts,
			&item.LastError,
			&item.CreatedAt,
			&item.PublishedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *OutboxRepo) MarkPublished(ctx context.Context, id uuid.UUID) error {
	const query = `
        UPDATE outbox_events
        SET status = 'PUBLISHED',
            published_at = NOW(),
            last_error = NULL
        WHERE id = $1
    `
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *OutboxRepo) MarkFailed(ctx context.Context, id uuid.UUID, lastError string) error {
	const query = `
        UPDATE outbox_events
        SET status = 'FAILED',
            attempts = attempts + 1,
            last_error = $2
        WHERE id = $1
    `
	_, err := r.db.Exec(ctx, query, id, lastError)
	return err
}
