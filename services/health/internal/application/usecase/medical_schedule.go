package usecase

import (
	"context"
	"health/internal/application/ports"
	"health/internal/domain/model"
	"strings"
	"time"

	"github.com/google/uuid"
)

type MedicalEntityReminderParams struct {
	PushEnabled         bool
	RemindOffsetMinutes *int
}

func validateMedicalEntityReminderParams(reminder *MedicalEntityReminderParams) error {
	if reminder == nil {
		return nil
	}
	_, err := validateReminderOffset(reminder.RemindOffsetMinutes)
	return err
}

func getMedicalEntityReminderSettings(ctx context.Context, repo ports.ScheduledRepository, petID uuid.UUID, sourceType string, sourceID uuid.UUID) (*MedicalEntityReminderParams, error) {
	item, err := repo.GetScheduledItemBySource(ctx, petID, sourceType, sourceID)
	if err != nil {
		if err == ports.ErrNotFound {
			return nil, nil
		}
		return nil, mapRepoErr(err)
	}
	var offset *int
	if item.RemindOffsetMinutes != nil {
		value := *item.RemindOffsetMinutes
		offset = &value
	}
	return &MedicalEntityReminderParams{
		PushEnabled:         item.PushEnabled,
		RemindOffsetMinutes: offset,
	}, nil
}

func generatedPlanReminder(requested, inherited *MedicalEntityReminderParams) *MedicalEntityReminderParams {
	if requested != nil {
		return requested
	}
	return inherited
}

func syncSystemOneShotScheduledItem(ctx context.Context, repo ports.ScheduledRepository, petID uuid.UUID, sourceType string, sourceID uuid.UUID, title string, note *string, startsAt *time.Time, shouldExist bool, userID uuid.UUID) (*model.ScheduledItem, error) {
	if !shouldExist || startsAt == nil || startsAt.IsZero() {
		return nil, mapRepoErr(repo.DeleteHealthScheduledItem(ctx, ports.DeleteHealthScheduledItemInput{
			PetID:           petID,
			SourceType:      sourceType,
			SourceID:        sourceID,
			DeletedByUserID: userID,
		}))
	}
	item, err := repo.UpsertHealthScheduledItem(ctx, ports.UpsertHealthScheduledItemInput{
		PetID:               petID,
		SourceType:          sourceType,
		SourceID:            sourceID,
		Title:               strings.TrimSpace(title),
		Note:                trimStringOrNil(note),
		StartsAt:            startsAt.UTC(),
		PushEnabled:         false,
		RemindOffsetMinutes: nil,
		CreatedByUserID:     userID,
		UpdatedByUserID:     userID,
	})
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if err := regenerateScheduledItemOccurrences(ctx, repo, item, time.Now().UTC()); err != nil {
		return nil, err
	}
	return item, nil
}

func applyMedicalEntityReminderSettings(ctx context.Context, repo ports.ScheduledRepository, petID uuid.UUID, reminder *MedicalEntityReminderParams, sourceType string, sourceID uuid.UUID, canExist bool, userID uuid.UUID) error {
	if reminder == nil {
		return nil
	}
	if !canExist {
		return ErrInvalidInput
	}
	item, err := repo.GetScheduledItemBySource(ctx, petID, sourceType, sourceID)
	if err != nil {
		return mapRepoErr(err)
	}
	remindOffsetMinutes, err := validateReminderOffset(reminder.RemindOffsetMinutes)
	if err != nil {
		return err
	}
	_, err = repo.UpdateScheduledItemReminderSettings(ctx, ports.UpdateScheduledItemReminderSettingsInput{
		ID:                  item.ID,
		PetID:               petID,
		RowVersion:          item.RowVersion,
		PushEnabled:         reminder.PushEnabled,
		RemindOffsetMinutes: remindOffsetMinutes,
		UpdatedBy:           userID,
	})
	return mapRepoErr(err)
}

func regenerateScheduledItemOccurrences(ctx context.Context, repo ports.ScheduledRepository, item *model.ScheduledItem, from time.Time) error {
	if item == nil {
		return ErrInvalidInput
	}
	if !from.IsZero() {
		if err := repo.DeleteScheduledItemOccurrencesFrom(ctx, ports.DeleteScheduledItemOccurrencesFromInput{
			ScheduledItemID: item.ID,
			From:            from.UTC(),
		}); err != nil {
			return mapRepoErr(err)
		}
	}
	return ensureScheduledItemOccurrences(ctx, repo, item, from.UTC())
}
