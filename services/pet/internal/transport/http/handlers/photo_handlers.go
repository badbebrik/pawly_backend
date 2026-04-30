package handlers

import (
	"net/http"
	"pet/internal/application/usecase"
	"pet/internal/transport/http/dto"

	"github.com/google/uuid"
)

func (h *Handlers) InitPetPhotoUpload(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	petID, ok := parseRouteUUID(w, r, "pet_id", "pet_id")
	if !ok {
		return
	}

	var req dto.InitPetPhotoUploadRequest
	if !decodeRequest(w, r, &req) {
		return
	}

	fileID, upload, err := h.useCases.InitPetPhotoUpload(r.Context(), usecase.InitPetPhotoUploadParams{
		UserID:            userID,
		PetID:             petID,
		MimeType:          req.MimeType,
		OriginalFilename:  req.OriginalFilename,
		ExpectedSizeBytes: req.ExpectedSizeBytes,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, dto.InitPetPhotoUploadResponse{
		FileID: fileID,
		Upload: uploadInfoToResponse(upload),
	})
}

func (h *Handlers) ConfirmPetPhotoUpload(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	petID, ok := parseRouteUUID(w, r, "pet_id", "pet_id")
	if !ok {
		return
	}

	var req dto.ConfirmPetPhotoUploadRequest
	if !decodeRequest(w, r, &req) {
		return
	}

	fileID, err := uuid.Parse(req.FileID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid file_id")
		return
	}

	pet, err := h.useCases.ConfirmPetPhotoUpload(r.Context(), usecase.ConfirmPetPhotoUploadParams{
		UserID:     userID,
		PetID:      petID,
		RowVersion: req.RowVersion,
		FileID:     fileID,
		SizeBytes:  req.SizeBytes,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.PetEnvelopeResponse{
		Pet: petToResponse(pet, h.getPetPhotoDownloadURL(r, pet)),
	})
}

func (h *Handlers) DeletePetPhoto(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	petID, ok := parseRouteUUID(w, r, "pet_id", "pet_id")
	if !ok {
		return
	}

	var req dto.DeletePetPhotoRequest
	if !decodeRequest(w, r, &req) {
		return
	}

	pet, err := h.useCases.DeletePetPhoto(r.Context(), usecase.DeletePetPhotoParams{
		UserID:     userID,
		PetID:      petID,
		RowVersion: req.RowVersion,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.PetEnvelopeResponse{
		Pet: petToResponse(pet, h.getPetPhotoDownloadURL(r, pet)),
	})
}
