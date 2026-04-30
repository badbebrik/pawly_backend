package app

import (
	"net/http"
	"pawly/pkg/httpjson"

	"github.com/go-chi/chi/v5"
	"profile/internal/transport/http/handlers"
	"profile/internal/transport/http/middleware"
)

func (a *App) setupRoutes() http.Handler {
	r := chi.NewRouter()

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		httpjson.Write(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	h := handlers.New(a.useCases)

	r.Group(func(r chi.Router) {
		r.Use(middleware.WithUserID)
		r.Get("/v1/profiles/me", h.GetMe)
		r.Patch("/v1/profiles/me", h.PatchMe)
		r.Patch("/v1/profiles/me/preferences", h.PatchPreferences)
		r.Post("/v1/profiles/me/avatar:init-upload", h.InitAvatarUpload)
		r.Post("/v1/profiles/me/avatar:confirm-upload", h.ConfirmAvatarUpload)
		r.Delete("/v1/profiles/me/avatar", h.DeleteAvatar)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.WithInternalToken(a.cfg.InternalServiceToken))
		r.Post("/internal/v1/profiles/users:batch-brief", h.BatchProfilesBrief)
	})

	return r
}
