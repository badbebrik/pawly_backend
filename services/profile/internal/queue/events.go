package queue

import (
	"github.com/google/uuid"
)

type UserCreatedEvent struct {
	Event  string         `json:"event"`
	UserID uuid.UUID      `json:"user_id"`
	Locale string         `json:"locale"`
	Data   map[string]any `json:"data"`
}
