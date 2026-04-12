package app

import (
	"chat/internal/transport/http/handlers"
	appmw "chat/internal/transport/http/middleware"
	"chat/internal/transport/ws"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

func (a *App) setupRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	h := handlers.New(
		a.useCases.OpenDirectConversation,
		a.useCases.ListConversations,
		a.useCases.GetConversation,
		a.useCases.GetUnreadSummary,
		a.useCases.GetMessageHistory,
		a.useCases.MarkRead,
	)
	wsHandler := ws.NewHandler(
		a.hub,
		a.rtPub,
		a.presence,
		time.Duration(a.cfg.PresenceTTL)*time.Second,
		time.Duration(a.cfg.PresenceHeartbeat)*time.Second,
		a.useCases.SendMessage,
		a.useCases.MarkRead,
		a.useCases.GetConversation,
		a.useCases.GetUnreadSummary,
	)

	withUser := appmw.WithUserID

	r.With(withUser).Get("/v1/chat/ws", wsHandler.ServeHTTP)

	r.Group(func(r chi.Router) {
		r.Use(withUser)
		r.Post("/v1/chat/conversations:open", h.OpenConversation)
		r.Get("/v1/chat/conversations", h.ListConversations)
		r.Get("/v1/chat/unread-summary", h.GetUnreadSummary)
		r.Get("/v1/chat/conversations/{conversation_id}", h.GetConversation)
		r.Get("/v1/chat/conversations/{conversation_id}/messages", h.GetMessageHistory)
		r.Post("/v1/chat/conversations/{conversation_id}/read", h.MarkRead)
	})

	return r
}
