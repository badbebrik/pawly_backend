package app

import (
	"auth/internal/transport/http/handlers"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func (a *App) setupRoutes() http.Handler {
	r := chi.NewRouter()

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`{"status":"ok"}`))
		if err != nil {
			return
		}
	})

	authHandlers := handlers.NewAuthHandlers(a.AuthSvc)

	r.Post("/auth/register/email", authHandlers.RegisterEmail)

	return r
}
