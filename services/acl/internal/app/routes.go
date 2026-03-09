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
	internal := handlers.NewInternalHandlers(a.aclSvc)

	r.Group(func(r chi.Router) {
		r.Use(appmw.WithUserID)

		r.Get("/v1/pets/{pet_id}/acl/roles", pub.ListRoles)
		r.Post("/v1/pets/{pet_id}/acl/roles", pub.CreateCustomRole)
		r.Delete("/v1/pets/{pet_id}/acl/roles/{role_id}", pub.DeleteCustomRole)

		r.Get("/v1/acl/presets", pub.ListPresets)

		r.Get("/v1/pets/{pet_id}/acl/members", pub.ListMembers)
		r.Get("/v1/pets/{pet_id}/acl/me", pub.GetMyAccess)
		r.Get("/v1/pets/{pet_id}/acl/bootstrap", pub.GetBootstrap)
		r.Patch("/v1/pets/{pet_id}/acl/members/{member_id}", pub.UpdateMemberPermissions)
		r.Delete("/v1/pets/{pet_id}/acl/members/{member_id}", pub.RemoveMember)

		r.Post("/v1/pets/{pet_id}/acl/invites", pub.CreateInvite)
		r.Get("/v1/pets/{pet_id}/acl/invites", pub.ListInvites)
		r.Delete("/v1/pets/{pet_id}/acl/invites/{invite_id}", pub.RevokeInvite)
		r.Post("/v1/acl/invites/accept-by-code", pub.AcceptInviteByCode)
		r.Post("/v1/acl/invites/accept-by-token", pub.AcceptInviteByToken)
	})

	r.Post("/v1/acl/invites/preview-by-token", pub.PreviewInviteByToken)

	r.Group(func(r chi.Router) {
		r.Use(appmw.WithInternalToken(a.cfg.InternalServiceToken))
		r.Post("/internal/v1/acl/is-member", internal.IsMember)
		r.Post("/internal/v1/acl/get-policy", internal.GetPolicy)
		r.Post("/internal/v1/acl/check", internal.Check)
		r.Post("/internal/v1/acl/list-pets-for-user", internal.ListPetsForUser)
	})

	return r
}
