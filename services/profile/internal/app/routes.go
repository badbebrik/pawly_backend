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
		r.Put("/v1/profile/me", hs.PutMe)
		r.Post("/v1/profile/me/avatar:init-upload", hs.InitAvatarUpload)
		r.Post("/v1/profile/me/avatar:confirm-upload", hs.ConfirmAvatarUpload)
		r.Post("/v1/profile/me/avatar:test-upload", hs.TestAvatarUpload)
	})

	r.Get("/internal/v1/profile/users/{user_id}/public-contact", hs.GetPublicContact)

	return r
}
