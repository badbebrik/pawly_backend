package app

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"profile/internal/service"
	"profile/internal/transport/http/handlers"
	"profile/internal/transport/http/middleware"
)

func (a *App) setupRoutes(svc *service.ProfileService) http.Handler {
	r := chi.NewRouter()

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`"status":"ok"`))
	})

	hs := handlers.NewHandlers(svc)

	r.Group(func(r chi.Router) {
		r.Use(middleware.WithUserID)
		r.Get("/profile/me", hs.GetMe)
		r.Patch("/profile/me", hs.PatchMe)
	})

	return r
}
