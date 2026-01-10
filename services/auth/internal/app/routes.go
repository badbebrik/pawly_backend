package app

import (
	"auth/internal/transport/http/handlers"
	appmw "auth/internal/transport/http/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
)

func (a *App) setupRoutes() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(appmw.WithLocale("ru"))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`{"status":"ok"}`))
		if err != nil {
			return
		}
	})
	authHandlers := handlers.NewAuthHandlers(a.AuthSvc)

	r.Post("/auth/register/email", authHandlers.RegisterEmail)
	r.Post("/auth/verify/email", authHandlers.VerifyEmail)

	return r
}
