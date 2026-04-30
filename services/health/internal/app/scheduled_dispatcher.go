package app

import (
	"context"
	healthuc "health/internal/application/usecase"
	"time"

	"github.com/rs/zerolog/log"
)

type scheduledDispatcher struct {
	dispatchUseCase *healthuc.ScheduledDispatcher
	scheduled       *healthuc.Scheduled
}

func newScheduledDispatcher(dispatchUseCase *healthuc.ScheduledDispatcher, scheduled *healthuc.Scheduled) *scheduledDispatcher {
	return &scheduledDispatcher{dispatchUseCase: dispatchUseCase, scheduled: scheduled}
}

func (d *scheduledDispatcher) Run(ctx context.Context, dispatchInterval time.Duration, dispatchBatchSize int, horizonInterval time.Duration, horizonBatchSize int) {
	if dispatchInterval <= 0 {
		dispatchInterval = time.Minute
	}
	if horizonInterval <= 0 {
		horizonInterval = time.Hour
	}
	dispatchTicker := time.NewTicker(dispatchInterval)
	defer dispatchTicker.Stop()
	horizonTicker := time.NewTicker(horizonInterval)
	defer horizonTicker.Stop()

	d.generateScheduledOccurrencesHorizon(ctx, horizonBatchSize)
	d.dispatchDueScheduledOccurrences(ctx, dispatchBatchSize)
	for {
		select {
		case <-ctx.Done():
			return
		case <-dispatchTicker.C:
			d.dispatchDueScheduledOccurrences(ctx, dispatchBatchSize)
		case <-horizonTicker.C:
			d.generateScheduledOccurrencesHorizon(ctx, horizonBatchSize)
		}
	}
}

func (d *scheduledDispatcher) dispatchDueScheduledOccurrences(ctx context.Context, batchSize int) {
	result, err := d.dispatchUseCase.DispatchDueScheduledOccurrences(ctx, batchSize)
	if err != nil {
		log.Error().Err(err).Msg("list due scheduled occurrences failed")
		return
	}
	for i := range result.Failures {
		failure := result.Failures[i]
		log.Error().
			Err(failure.Err).
			Str("operation", failure.Operation).
			Str("pet_id", failure.PetID.String()).
			Str("occurrence_id", failure.OccurrenceID.String()).
			Msg("scheduled dispatch step failed")
	}

	if result.Scanned > 0 {
		log.Info().
			Int("scanned", result.Scanned).
			Int("published", result.Published).
			Int("skipped", result.Skipped).
			Int("failed", result.Failed).
			Msg("scheduled dispatch tick finished")
	}
}

func (d *scheduledDispatcher) generateScheduledOccurrencesHorizon(ctx context.Context, batchSize int) {
	result, err := d.scheduled.GenerateOccurrencesHorizon(ctx, batchSize)
	if err != nil {
		log.Error().Err(err).Msg("generate scheduled occurrences horizon failed")
		return
	}
	if result.Scanned > 0 {
		log.Info().
			Int("scanned", result.Scanned).
			Int("failed", result.Failed).
			Msg("scheduled occurrences horizon tick finished")
	}
}
