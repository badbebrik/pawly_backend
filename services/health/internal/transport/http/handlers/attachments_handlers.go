package handlers

import (
	healthuc "health/internal/application/usecase"
	"health/internal/transport/http/dto"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *Handlers) InitAttachmentUpload(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}

	var req dto.InitAttachmentUploadRequest
	if !decodeBody(w, r, &req) {
		return
	}

	fileID, upload, err := h.logs.InitAttachmentUpload(r.Context(), healthuc.InitAttachmentUploadParams{
		UserID:            userID,
		PetID:             petID,
		EntityType:        req.EntityType,
		MimeType:          req.MimeType,
		OriginalFilename:  req.OriginalFilename,
		ExpectedSizeBytes: req.ExpectedSizeBytes,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, dto.InitAttachmentUploadResponse{
		FileID: fileID.String(),
		Upload: dto.UploadResponse{
			Method:    upload.Method,
			URL:       upload.URL,
			Headers:   upload.Headers,
			ExpiresAt: upload.ExpiresAt.UTC().Format(time.RFC3339),
		},
	})
}

func (h *Handlers) ConfirmAttachmentUpload(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}

	var req dto.ConfirmAttachmentUploadRequest
	if !decodeBody(w, r, &req) {
		return
	}

	fileID, err := uuid.Parse(req.FileID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid file_id")
		return
	}

	file, err := h.logs.ConfirmAttachmentUpload(r.Context(), healthuc.ConfirmAttachmentUploadParams{
		UserID:     userID,
		PetID:      petID,
		EntityType: req.EntityType,
		FileID:     fileID,
		SizeBytes:  req.SizeBytes,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.ConfirmAttachmentUploadResponse{
		File: dto.UploadedFileResponse{
			ID:               file.ID.String(),
			MimeType:         file.MimeType,
			SizeBytes:        file.SizeBytes,
			OriginalFilename: file.OriginalFilename,
		},
	})
}

func (h *Handlers) GetPetDocuments(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	cursor, err := decodeTimeCursor(r.URL.Query().Get("cursor"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid cursor")
		return
	}
	entityType := optionalQueryString(r, "entity_type")
	fileType := optionalQueryString(r, "file_type")
	resp, err := h.documents.ListPetDocuments(r.Context(), healthuc.ListPetDocumentsParams{
		UserID:     userID,
		PetID:      petID,
		Cursor:     cursor,
		Limit:      parseIntOrDefault(r.URL.Query().Get("limit"), 20),
		Q:          r.URL.Query().Get("q"),
		EntityType: entityType,
		FileType:   fileType,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	items := make([]dto.PetDocumentResponse, 0, len(resp.Items))
	for i := range resp.Items {
		items = append(items, petDocumentAppToDTO(resp.Items[i]))
	}
	writeJSON(w, http.StatusOK, dto.PetDocumentsListResponse{Items: items, NextCursor: encodeTimeCursor(resp.NextCursor)})
}

func (h *Handlers) RenamePetDocument(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	documentID, err := uuid.Parse(chi.URLParam(r, "document_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid document_id")
		return
	}
	var req dto.RenameDocumentRequest
	if !decodeBody(w, r, &req) {
		return
	}
	item, err := h.documents.RenamePetDocument(r.Context(), healthuc.RenamePetDocumentParams{
		UserID:     userID,
		PetID:      petID,
		DocumentID: documentID,
		FileName:   req.FileName,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, petDocumentAppToDTO(*item))
}
