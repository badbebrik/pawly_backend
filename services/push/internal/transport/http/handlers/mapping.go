package handlers

import (
	"push/internal/domain/model"
	"push/internal/transport/http/dto"
)

func deviceTokenToResponse(item *model.DeviceToken) dto.DeviceTokenResponse {
	return dto.DeviceTokenResponse{
		ID:        item.ID,
		UserID:    item.UserID,
		DeviceID:  item.DeviceID,
		Platform:  item.Platform,
		PushToken: item.PushToken,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}

func petPushSettingsToResponse(item *model.PetPushSettings) dto.PetPushSettingsResponse {
	return dto.PetPushSettingsResponse{
		PetID:                 item.PetID,
		ScheduledItemsEnabled: item.ScheduledItemsEnabled,
		CreatedAt:             item.CreatedAt,
		UpdatedAt:             item.UpdatedAt,
	}
}
