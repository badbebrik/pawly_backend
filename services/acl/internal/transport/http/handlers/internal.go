package handlers

import (
	acluc "acl/internal/application/usecase"
	"acl/internal/transport/http/dto"
	"net/http"

	"github.com/google/uuid"
)

type InternalHandlers struct {
	useCases *acluc.Set
}

func NewInternal(useCases *acluc.Set) *InternalHandlers {
	return &InternalHandlers{useCases: useCases}
}

func (h *InternalHandlers) IsMember(w http.ResponseWriter, r *http.Request) {
	var req dto.IsMemberRequest
	if !decodeInternalBody(w, r, &req) {
		return
	}
	petID, userID, ok := parsePetAndUserRaw(w, req.PetID, req.UserID)
	if !ok {
		return
	}
	isMember, err := h.useCases.IsMember(r.Context(), petID, userID)
	if err != nil {
		writeACLError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.IsMemberResponse{IsMember: isMember})
}

func (h *InternalHandlers) GetPolicy(w http.ResponseWriter, r *http.Request) {
	var req dto.GetPolicyRequest
	if !decodeInternalBody(w, r, &req) {
		return
	}
	petID, userID, ok := parsePetAndUserRaw(w, req.PetID, req.UserID)
	if !ok {
		return
	}
	res, err := h.useCases.GetPolicy(r.Context(), petID, userID)
	if err != nil {
		writeACLError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.GetPolicyResponse{
		MemberID: res.MemberID, Status: res.Status, IsPrimaryOwner: res.IsPrimaryOwner, Policy: res.Policy,
	})
}

func (h *InternalHandlers) Check(w http.ResponseWriter, r *http.Request) {
	var req dto.CheckRequest
	if !decodeInternalBody(w, r, &req) {
		return
	}
	petID, userID, ok := parsePetAndUserRaw(w, req.PetID, req.UserID)
	if !ok {
		return
	}
	allowed, err := h.useCases.Check(r.Context(), acluc.CheckParams{PetID: petID, UserID: userID, Action: req.Action})
	if err != nil {
		writeACLError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.CheckResponse{Allowed: allowed})
}

func (h *InternalHandlers) ListPetsForUser(w http.ResponseWriter, r *http.Request) {
	var req dto.ListPetsForUserRequest
	if !decodeInternalBody(w, r, &req) {
		return
	}
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid user id")
		return
	}
	memberships, err := h.useCases.ListPetMembershipsForUser(r.Context(), userID)
	if err != nil {
		writeACLError(w, err)
		return
	}
	petIDs := make([]string, 0, len(memberships))
	items := make([]dto.MemberResponse, 0, len(memberships))
	for i := range memberships {
		m := memberships[i]
		petIDs = append(petIDs, m.PetID.String())
		items = append(items, memberToResponse(&m, nil))
	}
	writeJSON(w, http.StatusOK, dto.ListPetsForUserResponse{PetIDs: petIDs, Memberships: items})
}

func (h *InternalHandlers) ListMembersForPet(w http.ResponseWriter, r *http.Request) {
	var req dto.ListMembersForPetRequest
	if !decodeInternalBody(w, r, &req) {
		return
	}
	petID, err := uuid.Parse(req.PetID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid pet id")
		return
	}
	members, err := h.useCases.ListActiveMembersForPet(r.Context(), petID)
	if err != nil {
		writeACLError(w, err)
		return
	}
	userIDs := make([]string, 0, len(members))
	items := make([]dto.MemberResponse, 0, len(members))
	for i := range members {
		m := members[i]
		userIDs = append(userIDs, m.UserID.String())
		items = append(items, memberToResponse(&m, nil))
	}
	writeJSON(w, http.StatusOK, dto.ListMembersForPetResponse{UserIDs: userIDs, Members: items})
}

func (h *InternalHandlers) TransferOwnership(w http.ResponseWriter, r *http.Request) {
	var req dto.TransferOwnershipRequest
	if !decodeInternalBody(w, r, &req) {
		return
	}
	petID, requesterUserID, ok := parsePetAndUserRaw(w, req.PetID, req.RequesterUserID)
	if !ok {
		return
	}
	targetMemberID, err := uuid.Parse(req.TargetMemberID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid target member id")
		return
	}
	res, err := h.useCases.TransferOwnership(r.Context(), acluc.TransferOwnershipParams{
		PetID: petID, RequesterID: requesterUserID, TargetMemberID: targetMemberID,
	})
	if err != nil {
		writeACLError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.TransferOwnershipResponse{
		PreviousOwnerMemberID: res.PreviousOwner.ID,
		PreviousOwnerUserID:   res.PreviousOwner.UserID,
		CurrentOwnerMemberID:  res.CurrentOwner.ID,
		CurrentOwnerUserID:    res.CurrentOwner.UserID,
	})
}

func decodeInternalBody(w http.ResponseWriter, r *http.Request, out any) bool {
	if err := decodeJSON(r, out); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid json")
		return false
	}
	return true
}

func parsePetAndUserRaw(w http.ResponseWriter, petRaw, userRaw string) (uuid.UUID, uuid.UUID, bool) {
	petID, err := uuid.Parse(petRaw)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid pet id")
		return uuid.Nil, uuid.Nil, false
	}
	userID, err := uuid.Parse(userRaw)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid user id")
		return uuid.Nil, uuid.Nil, false
	}
	return petID, userID, true
}
