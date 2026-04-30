package app

import (
	"auth/internal/transport/http/handlers"
	appmw "auth/internal/transport/http/middleware"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

func (a *App) setupRoutes() http.Handler {
	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(appmw.RequestLogger)
	r.Use(chimw.Recoverer)
	r.Use(appmw.WithLocale("ru"))

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	h := handlers.New(a.useCases)

	r.Post("/v1/auth/email:register", h.RegisterEmail)
	r.Post("/v1/auth/email-verification:resend", h.ResendEmailVerification)
	r.Post("/v1/auth/email-verification:verify", h.VerifyEmail)
	r.Post("/v1/auth/sessions:login-email", h.LoginEmail)
	r.Post("/v1/auth/sessions:login-oauth", h.LoginOAuth)
	r.Post("/v1/auth/sessions:revoke", h.Logout)
	r.Post("/v1/auth/sessions:revoke-all", h.LogoutAll)
	r.Post("/v1/auth/sessions:refresh", h.Refresh)
	r.Post("/v1/auth/password:change", h.ChangePassword)
	r.Post("/v1/auth/password-reset:request", h.PasswordResetRequest)
	r.Post("/v1/auth/password-reset:verify", h.PasswordResetVerify)
	r.Post("/v1/auth/password-reset:confirm", h.PasswordResetConfirm)

	return r
}
