package handlers

import (
	"acl/internal/application/ports"
	acluc "acl/internal/application/usecase"
	"acl/internal/transport/http/dto"
	appmw "acl/internal/transport/http/middleware"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type PublicHandlers struct {
	useCases           *acluc.Set
	profiles           ports.ProfileClient
	pets               ports.PetClient
	inviteDeeplinkBase string
}

func NewPublic(useCases *acluc.Set, profiles ports.ProfileClient, pets ports.PetClient, inviteDeeplinkBase string) *PublicHandlers {
	return &PublicHandlers{useCases: useCases, profiles: profiles, pets: pets, inviteDeeplinkBase: inviteDeeplinkBase}
}

func (h *PublicHandlers) NotImplemented(w http.ResponseWriter, _ *http.Request) {
	writeError(w, http.StatusNotImplemented, "not_implemented", "endpoint scaffolded, implementation pending")
}

func (h *PublicHandlers) GetMyAccess(w http.ResponseWriter, r *http.Request) {
	petID, userID, ok := parsePetAndUser(w, r)
	if !ok {
		return
	}
	member, err := h.useCases.GetMyAccess(r.Context(), petID, userID)
	if err != nil {
		writeACLError(w, err)
		return
	}
	profilesByUser := h.loadProfilesByUserID(r.Context(), []ports.MemberView{*member})
	writeJSON(w, http.StatusOK, dto.GetMyAccessResponse{
		PetID:  petID,
		Member: memberToResponse(member, profilesByUser[member.UserID]),
	})
}

func (h *PublicHandlers) LeavePet(w http.ResponseWriter, r *http.Request) {
	petID, userID, ok := parsePetAndUser(w, r)
	if !ok {
		return
	}
	member, err := h.useCases.LeavePet(r.Context(), petID, userID)
	if err != nil {
		writeACLError(w, err)
		return
	}
	profilesByUser := h.loadProfilesByUserID(r.Context(), []ports.MemberView{*member})
	writeJSON(w, http.StatusOK, dto.MemberEnvelopeResponse{
		Member: memberToResponse(member, profilesByUser[member.UserID]),
	})
}

func (h *PublicHandlers) GetBootstrap(w http.ResponseWriter, r *http.Request) {
	petID, userID, ok := parsePetAndUser(w, r)
	if !ok {
		return
	}
	res, err := h.useCases.GetBootstrap(r.Context(), petID, userID)
	if err != nil {
		writeACLError(w, err)
		return
	}
	profilesByUser := h.loadProfilesByUserID(r.Context(), append([]ports.MemberView{*res.Me}, res.Members...))
	writeJSON(w, http.StatusOK, bootstrapToResponse(petID, res, profilesByUser, h.inviteDeeplinkBase))
}

func (h *PublicHandlers) ListMembers(w http.ResponseWriter, r *http.Request) {
	petID, userID, ok := parsePetAndUser(w, r)
	if !ok {
		return
	}
	items, err := h.useCases.ListMembers(r.Context(), petID, userID)
	if err != nil {
		writeACLError(w, err)
		return
	}
	profilesByUser := h.loadProfilesByUserID(r.Context(), items)
	writeJSON(w, http.StatusOK, dto.MembersListResponse{Items: membersToResponse(items, profilesByUser)})
}

func (h *PublicHandlers) UpdateMemberPermissions(w http.ResponseWriter, r *http.Request) {
	petID, userID, ok := parsePetAndUser(w, r)
	if !ok {
		return
	}
	memberID, err := uuid.Parse(chi.URLParam(r, "member_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid member id")
		return
	}
	var req dto.UpdateMemberPermissionsRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid json")
		return
	}
	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid role id")
		return
	}
	member, err := h.useCases.UpdateMemberPermissions(r.Context(), acluc.UpdateMemberPermissionsParams{
		PetID: petID, RequesterID: userID, MemberID: memberID, RoleID: roleID, Policy: req.Policy,
	})
	if err != nil {
		writeACLError(w, err)
		return
	}
	profilesByUser := h.loadProfilesByUserID(r.Context(), []ports.MemberView{*member})
	writeJSON(w, http.StatusOK, dto.MemberEnvelopeResponse{Member: memberToResponse(member, profilesByUser[member.UserID])})
}

func (h *PublicHandlers) RemoveMember(w http.ResponseWriter, r *http.Request) {
	petID, userID, ok := parsePetAndUser(w, r)
	if !ok {
		return
	}
	memberID, err := uuid.Parse(chi.URLParam(r, "member_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid member id")
		return
	}
	member, err := h.useCases.RemoveMember(r.Context(), acluc.RemoveMemberParams{PetID: petID, RequesterID: userID, MemberID: memberID})
	if err != nil {
		writeACLError(w, err)
		return
	}
	profilesByUser := h.loadProfilesByUserID(r.Context(), []ports.MemberView{*member})
	writeJSON(w, http.StatusOK, dto.MemberEnvelopeResponse{Member: memberToResponse(member, profilesByUser[member.UserID])})
}

func (h *PublicHandlers) ListRoles(w http.ResponseWriter, r *http.Request) {
	petID, userID, ok := parsePetAndUser(w, r)
	if !ok {
		return
	}
	items, err := h.useCases.ListRoles(r.Context(), petID, userID)
	if err != nil {
		writeACLError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.RolesListResponse{Items: rolesToResponse(items)})
}

func (h *PublicHandlers) CreateCustomRole(w http.ResponseWriter, r *http.Request) {
	petID, userID, ok := parsePetAndUser(w, r)
	if !ok {
		return
	}
	var req dto.CreateCustomRoleRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid json")
		return
	}
	role, err := h.useCases.CreateCustomRole(r.Context(), acluc.CreateCustomRoleParams{PetID: petID, RequesterID: userID, Title: req.Title, Policy: req.Policy})
	if err != nil {
		writeACLError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, dto.RoleEnvelopeResponse{Role: roleToResponse(*role)})
}

func (h *PublicHandlers) UpdateCustomRole(w http.ResponseWriter, r *http.Request) {
	petID, userID, ok := parsePetAndUser(w, r)
	if !ok {
		return
	}
	roleID, err := uuid.Parse(chi.URLParam(r, "role_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid role id")
		return
	}
	var req dto.UpdateCustomRoleRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid json")
		return
	}
	role, err := h.useCases.UpdateCustomRole(r.Context(), acluc.UpdateCustomRoleParams{
		PetID: petID, RequesterID: userID, RoleID: roleID, Title: req.Title, Policy: req.Policy,
	})
	if err != nil {
		writeACLError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.RoleEnvelopeResponse{Role: roleToResponse(*role)})
}

func (h *PublicHandlers) DeleteCustomRole(w http.ResponseWriter, r *http.Request) {
	petID, userID, ok := parsePetAndUser(w, r)
	if !ok {
		return
	}
	roleID, err := uuid.Parse(chi.URLParam(r, "role_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid role id")
		return
	}
	if err := h.useCases.DeleteCustomRole(r.Context(), petID, userID, roleID); err != nil {
		writeACLError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *PublicHandlers) CreateInvite(w http.ResponseWriter, r *http.Request) {
	petID, userID, ok := parsePetAndUser(w, r)
	if !ok {
		return
	}
	var req dto.CreateInviteRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid json")
		return
	}
	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid role id")
		return
	}
	res, err := h.useCases.CreateInvite(r.Context(), acluc.CreateInviteParams{
		PetID: petID, CreatedByUserID: userID, RoleID: roleID, Policy: req.Policy,
	})
	if err != nil {
		writeACLError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, dto.InviteEnvelopeResponse{
		Invite: inviteToResponse(*res.Invite, res.DeeplinkURL),
	})
}

func (h *PublicHandlers) ListInvites(w http.ResponseWriter, r *http.Request) {
	petID, userID, ok := parsePetAndUser(w, r)
	if !ok {
		return
	}
	items, err := h.useCases.ListInvites(r.Context(), petID, userID)
	if err != nil {
		writeACLError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.InvitesListResponse{Items: invitesToResponseWithBase(items, h.inviteDeeplinkBase)})
}

func (h *PublicHandlers) RegenerateInviteLink(w http.ResponseWriter, r *http.Request) {
	petID, userID, ok := parsePetAndUser(w, r)
	if !ok {
		return
	}
	inviteID, err := uuid.Parse(chi.URLParam(r, "invite_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid invite id")
		return
	}
	res, err := h.useCases.RegenerateInviteLink(r.Context(), petID, userID, inviteID)
	if err != nil {
		writeACLError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.InviteEnvelopeResponse{
		Invite: inviteToResponse(*res.Invite, res.DeeplinkURL),
	})
}

func (h *PublicHandlers) RevokeInvite(w http.ResponseWriter, r *http.Request) {
	petID, userID, ok := parsePetAndUser(w, r)
	if !ok {
		return
	}
	inviteID, err := uuid.Parse(chi.URLParam(r, "invite_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid invite id")
		return
	}
	if err := h.useCases.RevokeInvite(r.Context(), petID, userID, inviteID); err != nil {
		writeACLError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *PublicHandlers) AcceptInviteByCode(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmw.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	var req dto.AcceptInviteByCodeRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid json")
		return
	}
	res, err := h.useCases.AcceptInviteByCode(r.Context(), req.Code, userID)
	if err != nil {
		writeACLError(w, err)
		return
	}
	profilesByUser := h.loadProfilesByUserID(r.Context(), []ports.MemberView{*res.Member})
	writeJSON(w, http.StatusOK, dto.AcceptInviteResponse{
		PetID:  res.PetID,
		Member: memberToResponse(res.Member, profilesByUser[res.Member.UserID]),
	})
}

func (h *PublicHandlers) AcceptInviteByToken(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmw.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	var req dto.AcceptInviteByTokenRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid json")
		return
	}
	res, err := h.useCases.AcceptInviteByToken(r.Context(), req.Token, userID)
	if err != nil {
		writeACLError(w, err)
		return
	}
	profilesByUser := h.loadProfilesByUserID(r.Context(), []ports.MemberView{*res.Member})
	writeJSON(w, http.StatusOK, dto.AcceptInviteResponse{
		PetID:  res.PetID,
		Member: memberToResponse(res.Member, profilesByUser[res.Member.UserID]),
	})
}

func (h *PublicHandlers) PreviewInviteByToken(w http.ResponseWriter, r *http.Request) {
	if _, ok := appmw.UserIDFromContext(r.Context()); !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	var req dto.PreviewInviteByTokenRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid json")
		return
	}
	invite, err := h.useCases.PreviewInviteByToken(r.Context(), req.Token)
	if err != nil {
		writeACLError(w, err)
		return
	}
	token := strings.TrimSpace(req.Token)
	pet := h.loadPetBrief(r.Context(), invite.PetID)
	writeJSON(w, http.StatusOK, dto.PreviewInviteResponse{
		Invite: inviteToResponse(*invite, h.inviteDeeplinkBase+token),
		Pet:    petBriefToResponse(pet),
	})
}

func parsePetAndUser(w http.ResponseWriter, r *http.Request) (uuid.UUID, uuid.UUID, bool) {
	petID, err := uuid.Parse(chi.URLParam(r, "pet_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid pet id")
		return uuid.Nil, uuid.Nil, false
	}
	userID, ok := appmw.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return uuid.Nil, uuid.Nil, false
	}
	return petID, userID, true
}
