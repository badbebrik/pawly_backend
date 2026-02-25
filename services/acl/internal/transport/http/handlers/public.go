package handlers

import (
	"acl/internal/model"
	"acl/internal/repository"
	"acl/internal/service"
	appmw "acl/internal/transport/http/middleware"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type PublicHandlers struct {
	svc *service.ACLService
}

type createInviteRequest struct {
	RoleID       string       `json:"role_id"`
	BasePresetID *string      `json:"base_preset_id"`
	Policy       model.Policy `json:"policy"`
}

type createCustomRoleRequest struct {
	Title string `json:"title"`
}

type updateMemberPermissionsRequest struct {
	RoleID       string       `json:"role_id"`
	BasePresetID *string      `json:"base_preset_id"`
	Policy       model.Policy `json:"policy"`
}

type acceptInviteByCodeRequest struct {
	Code string `json:"code"`
}

type acceptInviteByTokenRequest struct {
	Token string `json:"token"`
}

func NewPublicHandlers(svc *service.ACLService) *PublicHandlers {
	return &PublicHandlers{svc: svc}
}

func (h *PublicHandlers) NotImplemented(w http.ResponseWriter, _ *http.Request) {
	writeError(w, http.StatusNotImplemented, "NOT_IMPLEMENTED", "endpoint scaffolded, implementation pending")
}

func (h *PublicHandlers) GetMyAccess(w http.ResponseWriter, r *http.Request) {
	petID, userID, ok := parsePetAndUser(w, r)
	if !ok {
		return
	}

	member, err := h.svc.GetMyAccess(r.Context(), petID, userID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"pet_id": petID.String(),
		"member": memberToDTO(member),
	})
}

func (h *PublicHandlers) ListMembers(w http.ResponseWriter, r *http.Request) {
	petID, userID, ok := parsePetAndUser(w, r)
	if !ok {
		return
	}

	items, err := h.svc.ListMembers(r.Context(), petID, userID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	out := make([]any, 0, len(items))
	for i := range items {
		out = append(out, memberToDTO(&items[i]))
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": out})
}

func (h *PublicHandlers) UpdateMemberPermissions(w http.ResponseWriter, r *http.Request) {
	petID, userID, ok := parsePetAndUser(w, r)
	if !ok {
		return
	}

	memberID, err := uuid.Parse(chi.URLParam(r, "member_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid member_id")
		return
	}

	var req updateMemberPermissionsRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid body")
		return
	}

	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid role_id")
		return
	}

	var basePresetID *uuid.UUID
	if req.BasePresetID != nil && *req.BasePresetID != "" {
		id, err := uuid.Parse(*req.BasePresetID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid base_preset_id")
			return
		}
		basePresetID = &id
	}

	member, err := h.svc.UpdateMemberPermissions(r.Context(), service.UpdateMemberPermissionsParams{
		PetID:        petID,
		RequesterID:  userID,
		MemberID:     memberID,
		RoleID:       roleID,
		Policy:       req.Policy,
		BasePresetID: basePresetID,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"member": memberToDTO(member)})
}

func (h *PublicHandlers) RemoveMember(w http.ResponseWriter, r *http.Request) {
	petID, userID, ok := parsePetAndUser(w, r)
	if !ok {
		return
	}

	memberID, err := uuid.Parse(chi.URLParam(r, "member_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid member_id")
		return
	}

	member, err := h.svc.RemoveMember(r.Context(), service.RemoveMemberParams{
		PetID:       petID,
		RequesterID: userID,
		MemberID:    memberID,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"member": memberToDTO(member)})
}

func (h *PublicHandlers) ListRoles(w http.ResponseWriter, r *http.Request) {
	petID, userID, ok := parsePetAndUser(w, r)
	if !ok {
		return
	}

	items, err := h.svc.ListRoles(r.Context(), petID, userID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	out := make([]any, 0, len(items))
	for i := range items {
		out = append(out, roleToDTO(items[i]))
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": out})
}

func (h *PublicHandlers) CreateCustomRole(w http.ResponseWriter, r *http.Request) {
	petID, userID, ok := parsePetAndUser(w, r)
	if !ok {
		return
	}

	var req createCustomRoleRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid body")
		return
	}

	role, err := h.svc.CreateCustomRole(r.Context(), service.CreateCustomRoleParams{
		PetID:       petID,
		RequesterID: userID,
		Title:       req.Title,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"role": roleToDTO(*role)})
}

func (h *PublicHandlers) DeleteCustomRole(w http.ResponseWriter, r *http.Request) {
	petID, userID, ok := parsePetAndUser(w, r)
	if !ok {
		return
	}

	roleID, err := uuid.Parse(chi.URLParam(r, "role_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid role_id")
		return
	}

	if err := h.svc.DeleteCustomRole(r.Context(), petID, userID, roleID); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *PublicHandlers) ListPresets(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListPresets(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}

	out := make([]any, 0, len(items))
	for _, item := range items {
		var roleCode any
		if item.RoleCode == "" {
			roleCode = nil
		} else {
			roleCode = item.RoleCode
		}
		out = append(out, map[string]any{
			"id":        item.ID.String(),
			"name":      item.Name,
			"role_code": roleCode,
			"policy":    item.Policy,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": out})
}

func (h *PublicHandlers) CreateInvite(w http.ResponseWriter, r *http.Request) {
	petID, userID, ok := parsePetAndUser(w, r)
	if !ok {
		return
	}

	var req createInviteRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid body")
		return
	}

	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid role_id")
		return
	}

	var basePresetID *uuid.UUID
	if req.BasePresetID != nil && *req.BasePresetID != "" {
		id, err := uuid.Parse(*req.BasePresetID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid base_preset_id")
			return
		}
		basePresetID = &id
	}

	res, err := h.svc.CreateInvite(r.Context(), service.CreateInviteParams{
		PetID:           petID,
		CreatedByUserID: userID,
		RoleID:          roleID,
		Policy:          req.Policy,
		BasePresetID:    basePresetID,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"invite": inviteToDTO(*res.Invite, res.DeeplinkURL),
	})
}

func (h *PublicHandlers) ListInvites(w http.ResponseWriter, r *http.Request) {
	petID, userID, ok := parsePetAndUser(w, r)
	if !ok {
		return
	}

	items, err := h.svc.ListInvites(r.Context(), petID, userID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	out := make([]any, 0, len(items))
	for _, item := range items {
		out = append(out, inviteToDTO(item, ""))
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": out})
}

func (h *PublicHandlers) RevokeInvite(w http.ResponseWriter, r *http.Request) {
	petID, userID, ok := parsePetAndUser(w, r)
	if !ok {
		return
	}

	inviteID, err := uuid.Parse(chi.URLParam(r, "invite_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid invite_id")
		return
	}

	if err := h.svc.RevokeInvite(r.Context(), petID, userID, inviteID); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *PublicHandlers) AcceptInviteByCode(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmw.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user id")
		return
	}

	var req acceptInviteByCodeRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid body")
		return
	}

	res, err := h.svc.AcceptInviteByCode(r.Context(), req.Code, userID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"pet_id": res.PetID.String(),
		"member": memberToDTO(res.Member),
	})
}

func (h *PublicHandlers) AcceptInviteByToken(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmw.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user id")
		return
	}

	var req acceptInviteByTokenRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid body")
		return
	}

	res, err := h.svc.AcceptInviteByToken(r.Context(), req.Token, userID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"pet_id": res.PetID.String(),
		"member": memberToDTO(res.Member),
	})
}

func parsePetAndUser(w http.ResponseWriter, r *http.Request) (uuid.UUID, uuid.UUID, bool) {
	petID, err := uuid.Parse(chi.URLParam(r, "pet_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid pet_id")
		return uuid.Nil, uuid.Nil, false
	}

	userID, ok := appmw.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user id")
		return uuid.Nil, uuid.Nil, false
	}

	return petID, userID, true
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid input")
	case errors.Is(err, service.ErrForbidden):
		writeError(w, http.StatusForbidden, "FORBIDDEN", "forbidden")
	case errors.Is(err, service.ErrNotFound):
		writeError(w, http.StatusNotFound, "NOT_FOUND", "not found")
	case errors.Is(err, service.ErrConflict):
		writeError(w, http.StatusConflict, "CONFLICT", "conflict")
	default:
		writeError(w, http.StatusInternalServerError, "INTERNAL", "internal error")
	}
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{
		"error": map[string]any{
			"code":    code,
			"message": message,
		},
	})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func memberToDTO(m *repository.MemberView) map[string]any {
	return map[string]any{
		"id":               m.ID.String(),
		"pet_id":           m.PetID.String(),
		"user_id":          m.UserID.String(),
		"status":           m.Status,
		"is_primary_owner": m.IsPrimaryOwner,
		"role":             roleToDTO(m.Role),
		"policy":           m.Policy,
		"created_at":       m.CreatedAt.UTC().Format(time.RFC3339),
		"updated_at":       m.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func roleToDTO(r repository.RoleView) map[string]any {
	var code any
	if r.Code == "" {
		code = nil
	} else {
		code = r.Code
	}

	var petID any
	if r.PetID != nil {
		petID = r.PetID.String()
	}
	var createdBy any
	if r.CreatedByUserID != nil {
		createdBy = r.CreatedByUserID.String()
	}

	return map[string]any{
		"id":                 r.ID.String(),
		"kind":               r.Kind,
		"pet_id":             petID,
		"code":               code,
		"title":              r.Title,
		"created_by_user_id": createdBy,
	}
}

func inviteToDTO(inv repository.InviteView, deeplink string) map[string]any {
	var basePresetID any
	if inv.BasePresetID != nil {
		basePresetID = inv.BasePresetID.String()
	}

	var consumedAt any
	if inv.ConsumedAt != nil {
		consumedAt = inv.ConsumedAt.UTC().Format(time.RFC3339)
	}

	var consumedBy any
	if inv.ConsumedByUserID != nil {
		consumedBy = inv.ConsumedByUserID.String()
	}

	var deeplinkURL any
	if deeplink != "" {
		deeplinkURL = deeplink
	}

	return map[string]any{
		"id":                  inv.ID.String(),
		"pet_id":              inv.PetID.String(),
		"status":              inv.Status,
		"code":                inv.Code,
		"deeplink_url":        deeplinkURL,
		"expires_at":          inv.ExpiresAt.UTC().Format(time.RFC3339),
		"role":                roleToDTO(inv.Role),
		"base_preset_id":      basePresetID,
		"policy":              inv.Policy,
		"created_by_user_id":  inv.CreatedByUserID.String(),
		"created_at":          inv.CreatedAt.UTC().Format(time.RFC3339),
		"consumed_at":         consumedAt,
		"consumed_by_user_id": consumedBy,
	}
}
