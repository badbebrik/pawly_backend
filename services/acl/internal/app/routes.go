package app

import (
	"acl/internal/transport/http/handlers"
	appmw "acl/internal/transport/http/middleware"
	"net/http"
	"pawly/pkg/httpjson"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (a *App) setupRoutes() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		httpjson.Write(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	publicHandlers := handlers.NewPublic(a.useCases, a.profile, a.pet, a.cfg.InviteDeeplinkBase)
	internalHandlers := handlers.NewInternal(a.useCases)

	r.Group(func(r chi.Router) {
		r.Use(appmw.WithUserID)

		r.Get("/v1/pets/{pet_id}/acl/roles", publicHandlers.ListRoles)
		r.Post("/v1/pets/{pet_id}/acl/roles", publicHandlers.CreateCustomRole)
		r.Patch("/v1/pets/{pet_id}/acl/roles/{role_id}", publicHandlers.UpdateCustomRole)
		r.Delete("/v1/pets/{pet_id}/acl/roles/{role_id}", publicHandlers.DeleteCustomRole)

		r.Get("/v1/pets/{pet_id}/acl/members", publicHandlers.ListMembers)
		r.Get("/v1/pets/{pet_id}/acl/me", publicHandlers.GetMyAccess)
		r.Delete("/v1/pets/{pet_id}/acl/me", publicHandlers.LeavePet)
		r.Get("/v1/pets/{pet_id}/acl/bootstrap", publicHandlers.GetBootstrap)
		r.Patch("/v1/pets/{pet_id}/acl/members/{member_id}", publicHandlers.UpdateMemberPermissions)
		r.Delete("/v1/pets/{pet_id}/acl/members/{member_id}", publicHandlers.RemoveMember)

		r.Post("/v1/pets/{pet_id}/acl/invites", publicHandlers.CreateInvite)
		r.Get("/v1/pets/{pet_id}/acl/invites", publicHandlers.ListInvites)
		r.Post("/v1/pets/{pet_id}/acl/invites/{invite_id}:regenerate-link", publicHandlers.RegenerateInviteLink)
		r.Delete("/v1/pets/{pet_id}/acl/invites/{invite_id}", publicHandlers.RevokeInvite)
		r.Post("/v1/acl/invites:accept-by-code", publicHandlers.AcceptInviteByCode)
		r.Post("/v1/acl/invites:accept-by-token", publicHandlers.AcceptInviteByToken)
		r.Post("/v1/acl/invites:preview-by-token", publicHandlers.PreviewInviteByToken)
	})

	r.Group(func(r chi.Router) {
		r.Use(appmw.WithInternalToken(a.cfg.InternalServiceToken))
		r.Post("/internal/v1/acl:is-member", internalHandlers.IsMember)
		r.Post("/internal/v1/acl:get-policy", internalHandlers.GetPolicy)
		r.Post("/internal/v1/acl:check", internalHandlers.Check)
		r.Post("/internal/v1/acl:list-pets-for-user", internalHandlers.ListPetsForUser)
		r.Post("/internal/v1/acl:list-members-for-pet", internalHandlers.ListMembersForPet)
		r.Post("/internal/v1/acl:transfer-ownership", internalHandlers.TransferOwnership)
	})

	return r
}
