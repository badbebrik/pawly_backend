package outbox

import (
	"auth/internal/infrastructure/rabbit"
	"auth/internal/repository"
	"context"
	"encoding/json"
	"time"

	"github.com/rs/zerolog/log"
)

type EventPublisher interface {
	PublishEvent(ctx context.Context, ev rabbit.NotificationEvent) error
}

type Worker struct {
	repo      repository.OutboxRepository
	publisher EventPublisher
	interval  time.Duration
	batchSize int
}

func NewWorker(repo repository.OutboxRepository, publisher EventPublisher, interval time.Duration, batchSize int) *Worker {
	if interval <= 0 {
		interval = 2 * time.Second
	}
	if batchSize <= 0 {
		batchSize = 100
	}
	return &Worker{
		repo:      repo,
		publisher: publisher,
		interval:  interval,
		batchSize: batchSize,
	}
}

func (w *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	w.processBatch(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.processBatch(ctx)
		}
	}
}

func (w *Worker) processBatch(ctx context.Context) {
	items, err := w.repo.ListPending(ctx, w.batchSize)
	if err != nil {
		log.Error().Err(err).Msg("outbox list pending failed")
		return
	}

	for i := range items {
		var ev rabbit.NotificationEvent
		if err := json.Unmarshal(items[i].Payload, &ev); err != nil {
			log.Error().Err(err).Str("outbox_event_id", items[i].ID.String()).Msg("outbox payload decode failed")
			_ = w.repo.MarkFailed(ctx, items[i].ID, err.Error())
			continue
		}

		if err := w.publisher.PublishEvent(ctx, ev); err != nil {
			log.Error().Err(err).Str("outbox_event_id", items[i].ID.String()).Str("event", items[i].EventType).Msg("outbox publish failed")
			_ = w.repo.MarkFailed(ctx, items[i].ID, err.Error())
			continue
		}

		if err := w.repo.MarkPublished(ctx, items[i].ID); err != nil {
			log.Error().Err(err).Str("outbox_event_id", items[i].ID.String()).Msg("outbox mark published failed")
		}
	}
}
