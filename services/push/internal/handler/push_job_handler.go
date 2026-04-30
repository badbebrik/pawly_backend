package handler

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	pushuc "push/internal/application/usecase"
	"push/internal/domain/model"
	"push/internal/sender"

	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

type PushJobHandler struct {
	useCases *pushuc.Set
	sender   sender.Sender
}

func NewPushJobHandler(useCases *pushuc.Set, pushSender sender.Sender) *PushJobHandler {
	return &PushJobHandler{useCases: useCases, sender: pushSender}
}

func (h *PushJobHandler) Handle(ctx context.Context, msg amqp091.Delivery) {
	var job model.ScheduledOccurrencePushJob
	if err := json.Unmarshal(msg.Body, &job); err != nil {
		log.Warn().Err(err).Msg("invalid push job json")
		_ = msg.Nack(false, false)
		return
	}

	if strings.TrimSpace(job.Event) != "SCHEDULED_OCCURRENCE_DUE" {
		log.Info().Str("event", job.Event).Msg("unknown push job event, skipping")
		_ = msg.Ack(false)
		return
	}

	petID, err := uuid.Parse(strings.TrimSpace(job.PetID))
	if err != nil || petID == uuid.Nil {
		log.Warn().Err(err).Str("pet_id", job.PetID).Msg("invalid push job pet_id")
		_ = msg.Nack(false, false)
		return
	}

	sentCount := 0
	enabledUsers := 0
	for _, rawUserID := range job.UserIDs {
		userID, err := uuid.Parse(strings.TrimSpace(rawUserID))
		if err != nil || userID == uuid.Nil {
			continue
		}

		items, err := h.useCases.ListEligibleDeviceTokens(ctx, pushuc.ListEligibleDeviceTokensParams{
			UserID: userID,
			PetID:  petID,
		})
		if err != nil {
			log.Error().Err(err).Str("user_id", userID.String()).Str("pet_id", petID.String()).Msg("resolve eligible device tokens failed")
			_ = msg.Nack(false, true)
			return
		}
		if len(items) == 0 {
			continue
		}
		enabledUsers++
		for _, item := range items {
			if err := h.sender.Send(ctx, item, buildPushMessage(job)); err != nil {
				if errors.Is(err, sender.ErrInvalidDeviceToken) {
					log.Warn().
						Str("user_id", item.UserID.String()).
						Str("device_id", item.DeviceID).
						Msg("invalid push token, deleting device token")
					_ = h.useCases.DeleteDeviceToken(ctx, pushuc.DeleteDeviceTokenParams{
						UserID:   item.UserID,
						DeviceID: item.DeviceID,
					})
					continue
				}
				log.Error().
					Err(err).
					Str("user_id", item.UserID.String()).
					Str("device_id", item.DeviceID).
					Str("occurrence_id", job.OccurrenceID).
					Msg("send push failed")
				_ = msg.Nack(false, true)
				return
			}
			sentCount++
		}
	}

	log.Info().
		Str("occurrence_id", job.OccurrenceID).
		Str("pet_id", job.PetID).
		Int("users_total", len(job.UserIDs)).
		Int("users_enabled", enabledUsers).
		Int("sent_count", sentCount).
		Msg("processed scheduled occurrence push job")

	_ = msg.Ack(false)
}

func buildPushMessage(job model.ScheduledOccurrencePushJob) model.PushMessage {
	body := strings.TrimSpace(job.Note)
	if body == "" {
		body = "Запланированное действие для питомца"
	}

	return model.PushMessage{
		Title: strings.TrimSpace(job.Title),
		Body:  body,
		Data: map[string]string{
			"type":              "scheduled_occurrence",
			"occurrence_id":     strings.TrimSpace(job.OccurrenceID),
			"scheduled_item_id": strings.TrimSpace(job.ScheduledItemID),
			"pet_id":            strings.TrimSpace(job.PetID),
			"source_type":       strings.TrimSpace(job.SourceType),
			"scheduled_for":     strings.TrimSpace(job.ScheduledFor),
		},
	}
}
