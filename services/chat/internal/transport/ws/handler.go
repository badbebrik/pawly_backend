package ws

import (
	"chat/internal/app"
	"chat/internal/infrastructure/realtime"
)

type Handler struct {
	hub      *realtime.Hub
	useCases *app.UseCases
}

func NewHandler(hub *realtime.Hub, useCases *app.UseCases) *Handler {
	return &Handler{
		hub:      hub,
		useCases: useCases,
	}
}
