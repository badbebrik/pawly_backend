package handlers

import (
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
	return map[string]any{
		"id":    r.ID.String(),
		"kind":  r.Kind,
		"code":  code,
		"title": r.Title,
	}
}
