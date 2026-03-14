package app

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"profile/internal/service"
	"profile/internal/transport/http/handlers"
	"profile/internal/transport/http/middleware"
)

func (a *App) setupRoutes(svc *service.ProfileService) http.Handler {
	r := chi.NewRouter()

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	hs := handlers.NewHandlers(svc)

	r.Group(func(r chi.Router) {
		r.Use(middleware.WithUserID)
		r.Get("/v1/profile/me", hs.GetMe)
		r.Patch("/v1/profile/me", hs.PatchMe)
		r.Patch("/v1/profile/me/preferences", hs.PatchPreferences)
		r.Post("/v1/profile/me/avatar:init-upload", hs.InitAvatarUpload)
		r.Post("/v1/profile/me/avatar:confirm-upload", hs.ConfirmAvatarUpload)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.WithInternalToken(a.cfg.InternalServiceToken))
		r.Post("/internal/v1/profile/users:batch-brief", hs.BatchProfilesBrief)
	})

	return r
}
