package handlers

import (
	"net/http"
	"pet/internal/application/usecase"
	"pet/internal/transport/http/dto"
	"time"

	"github.com/google/uuid"
)

func (h *Handlers) ChangePetStatus(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	petID, ok := parseRouteUUID(w, r, "pet_id", "pet_id")
	if !ok {
		return
	}

	var req dto.ChangePetStatusRequest
	if !decodeRequest(w, r, &req) {
		return
	}

	missingSince, err := parseOptionalDate(req.MissingSince, time.RFC3339, "missing_since")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", err.Error())
		return
	}

	pet, err := h.useCases.ChangePetStatus(r.Context(), usecase.ChangePetStatusParams{
		UserID:       userID,
		PetID:        petID,
		RowVersion:   req.RowVersion,
		Status:       req.Status,
		MissingSince: missingSince,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.PetEnvelopeResponse{
		Pet: petToResponse(pet, h.getPetPhotoDownloadURL(r, pet)),
	})
}

func (h *Handlers) TransferOwnership(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	petID, ok := parseRouteUUID(w, r, "pet_id", "pet_id")
	if !ok {
		return
	}

	var req dto.TransferOwnershipRequest
	if !decodeRequest(w, r, &req) {
		return
	}

	targetMemberID, err := uuid.Parse(req.TargetMemberID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid target_member_id")
		return
	}

	pet, err := h.useCases.TransferOwnership(r.Context(), usecase.TransferPetOwnershipParams{
		UserID:         userID,
		PetID:          petID,
		RowVersion:     req.RowVersion,
		TargetMemberID: targetMemberID,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.PetEnvelopeResponse{
		Pet: petToResponse(pet, h.getPetPhotoDownloadURL(r, pet)),
	})
}
