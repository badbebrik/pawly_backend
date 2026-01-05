package model

import "github.com/google/uuid"

type NotificationEvent struct {
	Event    string         `json:"event"`
	UserID   uuid.UUID      `json:"user_id"`
	Locale   string         `json:"locale"`
	Channels []string       `json:"channels"`
	Data     map[string]any `json:"data"`
}
