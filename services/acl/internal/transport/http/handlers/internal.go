package handlers

import (
	"acl/internal/service"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

type InternalHandlers struct {
	svc *service.ACLService
}

type isMemberRequest struct {
	PetID  string `json:"pet_id"`
	UserID string `json:"user_id"`
}

type getPolicyRequest struct {
	PetID  string `json:"pet_id"`
	UserID string `json:"user_id"`
}

type checkRequest struct {
	PetID  string `json:"pet_id"`
	UserID string `json:"user_id"`
	Action string `json:"action"`
}

type listPetsForUserRequest struct {
	UserID string `json:"user_id"`
}

func NewInternalHandlers(svc *service.ACLService) *InternalHandlers {
	return &InternalHandlers{svc: svc}
}

func (h *InternalHandlers) IsMember(w http.ResponseWriter, r *http.Request) {
	var req isMemberRequest
	if !decodeInternalBody(w, r, &req) {
		return
	}

	petID, userID, ok := parsePetAndUserRaw(w, req.PetID, req.UserID)
	if !ok {
		return
	}

	isMember, err := h.svc.IsMember(r.Context(), petID, userID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"is_member": isMember})
}

func (h *InternalHandlers) GetPolicy(w http.ResponseWriter, r *http.Request) {
	var req getPolicyRequest
	if !decodeInternalBody(w, r, &req) {
		return
	}

	petID, userID, ok := parsePetAndUserRaw(w, req.PetID, req.UserID)
	if !ok {
		return
	}

	res, err := h.svc.GetPolicy(r.Context(), petID, userID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"member_id":        res.MemberID.String(),
		"status":           res.Status,
		"is_primary_owner": res.IsPrimaryOwner,
		"policy":           res.Policy,
	})
}

func (h *InternalHandlers) Check(w http.ResponseWriter, r *http.Request) {
	var req checkRequest
	if !decodeInternalBody(w, r, &req) {
		return
	}

	petID, userID, ok := parsePetAndUserRaw(w, req.PetID, req.UserID)
	if !ok {
		return
	}

	allowed, err := h.svc.Check(r.Context(), service.CheckParams{
		PetID:  petID,
		UserID: userID,
		Action: req.Action,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"allowed": allowed})
}

func (h *InternalHandlers) ListPetsForUser(w http.ResponseWriter, r *http.Request) {
	var req listPetsForUserRequest
	if !decodeInternalBody(w, r, &req) {
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid user_id")
		return
	}

	petIDs, err := h.svc.ListPetsForUser(r.Context(), userID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	items := make([]string, 0, len(petIDs))
	for _, id := range petIDs {
		items = append(items, id.String())
	}
	writeJSON(w, http.StatusOK, map[string]any{"pet_ids": items})
}

func decodeInternalBody(w http.ResponseWriter, r *http.Request, out any) bool {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(out); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid body")
		return false
	}
	return true
}

func parsePetAndUserRaw(w http.ResponseWriter, petRaw, userRaw string) (uuid.UUID, uuid.UUID, bool) {
	petID, err := uuid.Parse(petRaw)
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid pet_id")
		return uuid.Nil, uuid.Nil, false
	}
	userID, err := uuid.Parse(userRaw)
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid user_id")
		return uuid.Nil, uuid.Nil, false
	}
	return petID, userID, true
}
