package handlers

import (
	"net/http"

	"profile/internal/transport/http/dto"
	"profile/internal/transport/http/middleware"
)

func (h *Handlers) InitAvatarUpload(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	var req dto.InitAvatarUploadRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid json")
		return
	}

	fileID, upload, err := h.useCases.InitAvatarUpload.Execute(r.Context(), userID, req.MimeType, req.ExpectedSizeBytes)
	if err != nil {
		writeAvatarError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, dto.InitAvatarUploadResponse{
		FileID: fileID,
		Upload: uploadInfoToResponse(upload),
	})
}

func (h *Handlers) ConfirmAvatarUpload(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	var req dto.ConfirmAvatarUploadRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid json")
		return
	}

	profile, err := h.useCases.ConfirmAvatarUpload.Execute(r.Context(), userID, req.FileID, req.SizeBytes)
	if err != nil {
		writeAvatarError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.ConfirmAvatarUploadResponse{
		Profile: h.profileToResponse(r.Context(), profile),
	})
}

func (h *Handlers) DeleteAvatar(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	profile, err := h.useCases.DeleteAvatar.Execute(r.Context(), userID)
	if err != nil {
		writeAvatarError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.ConfirmAvatarUploadResponse{
		Profile: h.profileToResponse(r.Context(), profile),
	})
}
