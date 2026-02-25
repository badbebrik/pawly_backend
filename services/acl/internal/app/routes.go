package app

import (
	"acl/internal/transport/http/handlers"
	appmw "acl/internal/transport/http/middleware"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (a *App) setupRoutes() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	pub := handlers.NewPublicHandlers(a.aclSvc)
	internal := handlers.NewInternalHandlers()

	r.Group(func(r chi.Router) {
		r.Use(appmw.WithUserID)

		r.Get("/v1/pets/{pet_id}/acl/roles", pub.ListRoles)
		r.Post("/v1/pets/{pet_id}/acl/roles", pub.NotImplemented)
		r.Delete("/v1/pets/{pet_id}/acl/roles/{role_id}", pub.NotImplemented)

		r.Get("/v1/acl/presets", pub.NotImplemented)

		r.Get("/v1/pets/{pet_id}/acl/members", pub.ListMembers)
		r.Get("/v1/pets/{pet_id}/acl/me", pub.GetMyAccess)
		r.Patch("/v1/pets/{pet_id}/acl/members/{member_id}", pub.NotImplemented)
		r.Delete("/v1/pets/{pet_id}/acl/members/{member_id}", pub.NotImplemented)

		r.Post("/v1/pets/{pet_id}/acl/invites", pub.NotImplemented)
		r.Get("/v1/pets/{pet_id}/acl/invites", pub.NotImplemented)
		r.Delete("/v1/pets/{pet_id}/acl/invites/{invite_id}", pub.NotImplemented)
		r.Post("/v1/acl/invites/accept-by-code", pub.NotImplemented)
		r.Post("/v1/acl/invites/accept-by-token", pub.NotImplemented)
	})

	r.Group(func(r chi.Router) {
		r.Use(appmw.WithInternalToken(a.cfg.InternalServiceToken))
		r.Post("/internal/v1/acl/is-member", internal.NotImplemented)
		r.Post("/internal/v1/acl/get-policy", internal.NotImplemented)
		r.Post("/internal/v1/acl/check", internal.NotImplemented)
		r.Post("/internal/v1/acl/list-pets-for-user", internal.NotImplemented)
	})

	return r
}
